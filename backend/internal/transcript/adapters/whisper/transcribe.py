#!/usr/bin/env python3
"""
Recognition engine entry point. Called as a subprocess by the Go adapter (plan P-4).

Contract:
  stdin  — not used
  stdout — one JSON object (see scriptOutput in engine.go)
  stderr — diagnostics only, NEVER recognised text
  exit 0 — stdout carries a valid result
  exit 1 — recognition failed; stderr carries the reason

Arguments:
  --model PATH    path to the model weights directory
  --audio PATH    path to the 16 kHz mono WAV to recognise
  --language TAG  optional BCP-47 hint; absent means auto-detect
"""

import argparse
import json
import math
import sys
import time
import uuid


def main() -> None:
    parser = argparse.ArgumentParser(description="ASR engine (faster-whisper)")
    parser.add_argument("--model", required=True, help="path to model weights")
    parser.add_argument("--audio", required=True, help="path to 16 kHz mono audio")
    parser.add_argument("--language", default=None, help="BCP-47 language hint")
    args = parser.parse_args()

    try:
        from faster_whisper import WhisperModel  # type: ignore[import-untyped]
    except ImportError:
        print("faster-whisper is not installed", file=sys.stderr)
        sys.exit(1)

    t0 = time.monotonic()
    model = WhisperModel(
        args.model,
        device="cpu",
        compute_type="int8",
    )
    model_load_sec = time.monotonic() - t0
    print(f"model loaded in {model_load_sec:.1f}s", file=sys.stderr)

    t1 = time.monotonic()
    raw_segments, info = model.transcribe(
        args.audio,
        language=args.language,
        vad_filter=True,         # mandatory — §4.5, A01
        vad_parameters=dict(
            threshold=0.5,       # default Silero threshold, not swept
        ),
        beam_size=5,
        word_timestamps=False,
    )

    segments = []
    for seg in raw_segments:
        # avg_logprob is a log-probability (negative), not a 0-1 confidence. The domain
        # rejects values outside [0, 1], so we convert via exp(). A typical value of -0.3
        # becomes ~0.74; values near 0 approach 1.0 (high confidence).
        confidence = None
        if seg.avg_logprob is not None:
            confidence = round(min(1.0, max(0.0, math.exp(seg.avg_logprob))), 4)

        segments.append({
            "id": str(uuid.uuid4()),
            "start_ms": round(seg.start * 1000),
            "end_ms": round(seg.end * 1000),
            "text": seg.text.strip(),
            "confidence": confidence,
        })

    transcribe_sec = time.monotonic() - t1
    print(f"transcribed in {transcribe_sec:.1f}s", file=sys.stderr)

    # speech_seconds is the total audio AFTER VAD removed non-speech. It is the signal
    # that decides `no_speech` (P-4a): duration_after_vad < 1s → no_speech.
    speech_seconds = info.duration_after_vad if info.duration_after_vad is not None else info.duration

    result = {
        "language": info.language or "",
        "speech_seconds": round(speech_seconds, 3),
        "segments": segments,
        "model_load_sec": round(model_load_sec, 3),
        "transcribe_sec": round(transcribe_sec, 3),
    }

    # stdout is the ONLY channel for structured output. stderr is for diagnostics.
    json.dump(result, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
