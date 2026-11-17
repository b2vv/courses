"""Stdlib-only stand-in for the real transcribe.py boundary test."""

import json
import os
from types import SimpleNamespace


class WhisperModel:
    def __init__(self, model, **kwargs):
        pass

    def transcribe(self, audio, **kwargs):
        with open(os.environ["WHISPER_TEST_KWARGS"], "w", encoding="utf-8") as output:
            json.dump(kwargs, output)
        return [], SimpleNamespace(language="en", duration=2.0, duration_after_vad=2.0)
