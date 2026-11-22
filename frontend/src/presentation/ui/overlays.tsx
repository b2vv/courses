import type { ReactElement, ReactNode } from 'react';
import * as Aria from 'react-aria-components';
import { useTranslation } from 'react-i18next';
import { Button } from './controls';

type DialogProps = {
  triggerLabel: string;
  /** Styles the default trigger button; a primary trigger marks the page's main call to action.
   * Ignored when a composed `trigger` is supplied. Defaults to `secondary`. */
  triggerVariant?: 'primary' | 'secondary';
  /** Optional composed trigger, e.g. a tooltip around an icon button. */
  trigger?: ReactElement;
  /** When set, the trigger keeps `triggerLabel` as its visible text but exposes this as its
   * accessible name — lets a dense table show a short verb while assistive tech and tests still
   * read the full, name-qualified action. */
  triggerAriaLabel?: string;
  title: string;
  closeLabel: string;
  children: ReactNode | ((close: () => void) => ReactNode);
};

export function Dialog({
  triggerLabel,
  triggerVariant = 'secondary',
  trigger,
  triggerAriaLabel,
  title,
  closeLabel,
  children,
}: DialogProps) {
  return (
    <Aria.DialogTrigger>
      {trigger ?? (
        <Button
          variant={triggerVariant}
          {...(triggerAriaLabel ? { 'aria-label': triggerAriaLabel } : {})}
        >
          {triggerLabel}
        </Button>
      )}
      <Aria.ModalOverlay className="ui-modal-overlay" isDismissable>
        <Aria.Modal className="ui-modal">
          <Aria.Dialog className="ui-dialog">
            {({ close }) => (
              <>
                <Aria.Heading slot="title">{title}</Aria.Heading>
                {typeof children === 'function' ? children(close) : children}
                <div className="ui-dialog-actions">
                  <Button variant="secondary" onPress={close}>
                    {closeLabel}
                  </Button>
                </div>
              </>
            )}
          </Aria.Dialog>
        </Aria.Modal>
      </Aria.ModalOverlay>
    </Aria.DialogTrigger>
  );
}

type MenuCardProps = {
  /** Visible text of the trigger button, e.g. "Menu". */
  triggerLabel: string;
  /** Dialog heading. */
  title: string;
  /** Accessible name of the icon-only close button. */
  closeLabel: string;
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  children: ReactNode;
};

/**
 * A menu that opens as a card pinned to the top right of the viewport, sized to its content
 * (owner 2026-11-22: a card fixed to the right with a 32 px margin, not a sliding sheet). It is
 * a modal dialog: focus is held inside, Escape or a press outside closes it, and focus returns
 * to the trigger. The state is controlled so the caller can close it on navigation.
 */
export function MenuCard({
  triggerLabel,
  title,
  closeLabel,
  isOpen,
  onOpenChange,
  children,
}: MenuCardProps) {
  return (
    <Aria.DialogTrigger isOpen={isOpen} onOpenChange={onOpenChange}>
      <Button variant="secondary" className="ui-menu-card-trigger">
        <svg viewBox="0 0 20 20" aria-hidden="true" focusable="false">
          <path d="M3 5h14M3 10h14M3 15h14" />
        </svg>
        {triggerLabel}
      </Button>
      <Aria.ModalOverlay className="ui-menu-card-overlay" isDismissable>
        <Aria.Modal className="ui-menu-card">
          <Aria.Dialog className="ui-menu-card-dialog">
            <div className="ui-menu-card-head">
              <Aria.Heading slot="title">{title}</Aria.Heading>
              <Button variant="secondary" slot="close" aria-label={closeLabel}>
                <svg viewBox="0 0 20 20" aria-hidden="true" focusable="false">
                  <path d="M5 5l10 10M15 5L5 15" />
                </svg>
              </Button>
            </div>
            {children}
          </Aria.Dialog>
        </Aria.Modal>
      </Aria.ModalOverlay>
    </Aria.DialogTrigger>
  );
}

export type MenuOption = { id: string; label: string; isDisabled?: boolean };

type MenuProps = {
  label: string;
  items: readonly MenuOption[];
  onAction: (key: Aria.Key) => void;
};

export function Menu({ label, items, onAction }: MenuProps) {
  return (
    <Aria.MenuTrigger>
      <Button variant="secondary">{label}</Button>
      <Aria.Popover className="ui-popover ui-popover-list">
        <Aria.Menu items={items} onAction={onAction} className="ui-listbox">
          {(item) => (
            <Aria.MenuItem
              id={item.id}
              textValue={item.label}
              isDisabled={item.isDisabled ?? false}
              className="ui-option"
            >
              {item.label}
            </Aria.MenuItem>
          )}
        </Aria.Menu>
      </Aria.Popover>
    </Aria.MenuTrigger>
  );
}

export type TabItem = { id: string; label: string; content: ReactNode };

type TabsProps = Omit<Aria.TabsProps, 'children' | 'className'> & {
  label: string;
  items: readonly TabItem[];
};

export function Tabs({ label, items, ...props }: TabsProps) {
  return (
    <Aria.Tabs {...props} className="ui-tabs">
      <Aria.TabList aria-label={label} items={items} className="ui-tab-list">
        {(item) => (
          <Aria.Tab id={item.id} className="ui-tab">
            {item.label}
          </Aria.Tab>
        )}
      </Aria.TabList>
      {items.map((item) => (
        <Aria.TabPanel id={item.id} key={item.id} className="ui-tab-panel">
          {item.content}
        </Aria.TabPanel>
      ))}
    </Aria.Tabs>
  );
}

type TooltipProps = { children: ReactElement; content: string };

export function Tooltip({ children, content }: TooltipProps) {
  return (
    <Aria.TooltipTrigger delay={500}>
      {children}
      <Aria.Tooltip className="ui-tooltip">{content}</Aria.Tooltip>
    </Aria.TooltipTrigger>
  );
}

type ToastProps = { message: string; onDismiss: () => void };

export function Toast({ message, onDismiss }: ToastProps) {
  const { t } = useTranslation();
  return (
    <div className="ui-toast">
      <output aria-live="polite" aria-atomic="true">
        {message}
      </output>
      <Button variant="secondary" onPress={onDismiss} aria-label={t('ui.closeMessage')}>
        {t('common.close')}
      </Button>
    </div>
  );
}
