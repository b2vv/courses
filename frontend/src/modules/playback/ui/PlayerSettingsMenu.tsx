import * as Aria from 'react-aria-components';
import { RadioGroup, Switch, type ChoiceOption } from '../../../presentation/ui';

type Group = {
  readonly label: string;
  readonly value: string;
  readonly options: readonly ChoiceOption[];
  readonly onChange: (value: string) => void;
};

type PlayerSettingsMenuProps = {
  /** Names the gear button and its panel for assistive tech. */
  readonly label: string;
  readonly speed: Group;
  /** Present only for adaptive (HLS) sources; a native file has one rendition and no choice. */
  readonly quality?: Group;
  /** Present only when the lesson has a transcript to show as captions. */
  readonly captions?: {
    readonly label: string;
    readonly isOn: boolean;
    readonly onChange: (on: boolean) => void;
  };
  /** Present only when a next lesson exists to advance to; durable autoplay-next preference. */
  readonly autoplay?: {
    readonly label: string;
    readonly isOn: boolean;
    readonly onChange: (on: boolean) => void;
  };
};

/**
 * One gear that consolidates speed, quality and captions, the way YouTube/Hulu fold their settings
 * into a single button. It is a popover of real radio groups and a switch — keyboard- and
 * screen-reader-navigable — rather than inline selects crowding the control bar. Closed by default,
 * so the bar carries just the gear and the layout stays simple at phone width and 200% zoom.
 */
export function PlayerSettingsMenu({
  label,
  speed,
  quality,
  captions,
  autoplay,
}: PlayerSettingsMenuProps) {
  return (
    <Aria.DialogTrigger>
      <Aria.Button className="lesson-player-iconbtn" aria-label={label}>
        <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none">
          <circle cx="12" cy="12" r="3.2" stroke="currentColor" strokeWidth="1.6" />
          <path
            d="M12 2.5v3M12 18.5v3M21.5 12h-3M5.5 12h-3M18.7 5.3l-2.1 2.1M7.4 16.6l-2.1 2.1M18.7 18.7l-2.1-2.1M7.4 7.4 5.3 5.3"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinecap="round"
          />
        </svg>
      </Aria.Button>
      <Aria.Popover className="ui-popover lesson-player-settings">
        <Aria.Dialog aria-label={label} className="lesson-player-settings-panel">
          <RadioGroup
            label={speed.label}
            value={speed.value}
            onChange={speed.onChange}
            options={speed.options}
          />
          {quality === undefined ? null : (
            <RadioGroup
              label={quality.label}
              value={quality.value}
              onChange={quality.onChange}
              options={quality.options}
            />
          )}
          {captions === undefined ? null : (
            <Switch isSelected={captions.isOn} onChange={captions.onChange}>
              {captions.label}
            </Switch>
          )}
          {autoplay === undefined ? null : (
            <Switch isSelected={autoplay.isOn} onChange={autoplay.onChange}>
              {autoplay.label}
            </Switch>
          )}
        </Aria.Dialog>
      </Aria.Popover>
    </Aria.DialogTrigger>
  );
}
