import type { ReactElement, ReactNode } from 'react';
import * as Aria from 'react-aria-components';
import { useTranslation } from 'react-i18next';
import { Button } from './controls';

type DialogProps = {
  triggerLabel: string;
  title: string;
  children: ReactNode;
};

export function Dialog({ triggerLabel, title, children }: DialogProps) {
  return (
    <Aria.DialogTrigger>
      <Button variant="secondary">{triggerLabel}</Button>
      <Aria.ModalOverlay className="ui-modal-overlay" isDismissable>
        <Aria.Modal className="ui-modal">
          <Aria.Dialog className="ui-dialog">
            {({ close }) => (
              <>
                <Aria.Heading slot="title">{title}</Aria.Heading>
                {children}
                <div className="ui-dialog-actions">
                  <Button variant="secondary" onPress={close}>
                    Закрити
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
      <Aria.Popover className="ui-popover">
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
