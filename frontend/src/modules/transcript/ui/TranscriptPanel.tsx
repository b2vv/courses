import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  activeSegmentIndex,
  documentsOf,
  searchSegments,
  type LessonTranscript,
  type TranscriptSegment,
} from '../domain/transcript';
import { formatTimestamp } from '../../../presentation/format-timestamp';

type TranscriptPanelProps = {
  readonly transcript: LessonTranscript;
  readonly positionMs: number;
  /** Undefined while the player cannot seek — before the first play the duration is unknown, and
   * a click would be clamped into the first second instead of going where it says. */
  readonly onSeek?: (positionMs: number) => void;
};

export function TranscriptPanel({ transcript, positionMs, onSeek }: TranscriptPanelProps) {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const [following, setFollowing] = useState(true);
  const activeRef = useRef<HTMLLIElement>(null);

  const matches = useMemo(
    () => searchSegments(transcript.segments, query),
    [transcript.segments, query],
  );
  const searching = query.trim() !== '';
  const shown: readonly TranscriptSegment[] = searching ? matches : transcript.segments;
  const active = activeSegmentIndex(transcript.segments, positionMs);
  const activeId = active === -1 ? null : (transcript.segments[active]?.id ?? null);

  useEffect(() => {
    if (!following || searching || activeRef.current === null) return;
    activeRef.current.scrollIntoView({ block: 'nearest' });
  }, [activeId, following, searching]);

  if (transcript.language === null || transcript.segments.length === 0) {
    return (
      <section className="transcript" aria-labelledby="transcript-heading">
        <h2 id="transcript-heading">{t('learn.transcript.heading')}</h2>
        <p className="transcript-empty">{t('learn.transcript.none')}</p>
      </section>
    );
  }

  return (
    <section className="transcript" aria-labelledby="transcript-heading">
      <header className="transcript-header">
        <h2 id="transcript-heading">{t('learn.transcript.heading')}</h2>
        {transcript.stale ? (
          <output className="transcript-stale">{t('learn.transcript.stale')}</output>
        ) : null}
        <label className="transcript-search">
          <span>{t('learn.transcript.search')}</span>
          <input
            type="search"
            value={query}
            placeholder={t('learn.transcript.searchPlaceholder')}
            onChange={(event) => setQuery(event.target.value)}
          />
        </label>
        {searching ? (
          <output className="transcript-matches">
            {matches.length === 0
              ? t('learn.transcript.noMatches')
              : t('learn.transcript.matches', { count: matches.length })}
          </output>
        ) : null}
        {!following && !searching ? (
          <button type="button" onClick={() => setFollowing(true)}>
            {t('learn.transcript.follow')}
          </button>
        ) : null}
      </header>

      {/*
        Following is paused by USER INTENT, not by a scroll event. Autoscroll moves the list and
        fires the very same scroll event a person does, so distinguishing them by a timer means
        any real scroll landing in the same tick is swallowed — and the window where that happens
        is invisible. Wheel, touch and drag never come from scrollIntoView, so this needs no
        guessing at all.
      */}
      <ol
        className="transcript-lines"
        aria-label={t('learn.transcript.heading')}
        onWheel={() => setFollowing(false)}
        onTouchMove={() => setFollowing(false)}
        onPointerDown={() => setFollowing(false)}
      >
        {shown.map((segment) => {
          const isActive = segment.id === activeId;
          const documents = documentsOf(transcript, segment);
          return (
            <li
              key={segment.id}
              ref={isActive ? activeRef : undefined}
              className={isActive ? 'transcript-line transcript-line-active' : 'transcript-line'}
              aria-current={isActive ? 'true' : undefined}
            >
              {/*
                A line is a BUTTON only while seeking is possible. A disabled button is removed
                from keyboard traversal, so rendering one would make the text itself unreachable
                for anyone not using a mouse — the transcript exists precisely for those readers.
                Without a seek the same content is plain text: readable, reachable, and not a dead
                control that invites a click which does nothing.
              */}
              {onSeek === undefined ? (
                <p className="transcript-line-body">
                  <time>{formatTimestamp(segment.startMs)}</time>
                  {segment.speaker === undefined ? null : (
                    <span className="transcript-speaker">{segment.speaker}</span>
                  )}
                  <span className="transcript-text">{segment.text}</span>
                </p>
              ) : (
                <button
                  type="button"
                  className="transcript-line-body"
                  aria-label={t('learn.transcript.seek', {
                    time: formatTimestamp(segment.startMs),
                  })}
                  onClick={() => onSeek(segment.startMs)}
                >
                  <time>{formatTimestamp(segment.startMs)}</time>
                  {segment.speaker === undefined ? null : (
                    <span className="transcript-speaker">{segment.speaker}</span>
                  )}
                  <span className="transcript-text">{segment.text}</span>
                </button>
              )}
              {documents.length === 0 ? null : (
                <ul className="transcript-documents" aria-label={t('learn.transcript.documents')}>
                  {documents.map((document) => (
                    <li key={document.id}>
                      <a href={document.url} target="_blank" rel="noreferrer noopener">
                        {document.label}
                      </a>
                    </li>
                  ))}
                </ul>
              )}
            </li>
          );
        })}
      </ol>
    </section>
  );
}
