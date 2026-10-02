import type { ReactNode } from 'react';
import * as Aria from 'react-aria-components';

type ButtonProps = Omit<Aria.ButtonProps, 'className'> & {
  variant?: 'primary' | 'secondary' | 'danger';
};

export function Button({ variant = 'primary', ...props }: ButtonProps) {
  return <Aria.Button {...props} className={`ui-button ui-button--${variant}`} />;
}

export function Link(props: Aria.LinkProps) {
  return (
    <Aria.Link
      {...props}
      className={Aria.composeRenderProps(props.className, (name) => `ui-link ${name ?? ''}`)}
    />
  );
}

type FieldProps = Omit<Aria.TextFieldProps, 'children' | 'className'> & {
  label: string;
  description?: string;
  errorMessage?: string;
  placeholder?: string;
};

function FieldHelp({
  description,
  errorMessage,
}: {
  description: string | undefined;
  errorMessage: string | undefined;
}) {
  return (
    <>
      {description && (
        <Aria.Text slot="description" className="ui-description">
          {description}
        </Aria.Text>
      )}
      <Aria.FieldError className="ui-field-error">{errorMessage}</Aria.FieldError>
    </>
  );
}

export function TextField({ label, description, errorMessage, placeholder, ...props }: FieldProps) {
  return (
    <Aria.TextField {...props} className="ui-field">
      <Aria.Label>{label}</Aria.Label>
      <Aria.Input {...(placeholder === undefined ? {} : { placeholder })} className="ui-input" />
      <FieldHelp description={description} errorMessage={errorMessage} />
    </Aria.TextField>
  );
}

export function TextArea({ label, description, errorMessage, placeholder, ...props }: FieldProps) {
  return (
    <Aria.TextField {...props} className="ui-field">
      <Aria.Label>{label}</Aria.Label>
      <Aria.TextArea placeholder={placeholder} className="ui-input ui-textarea" />
      <FieldHelp description={description} errorMessage={errorMessage} />
    </Aria.TextField>
  );
}

export type ChoiceOption = { id: string; label: string };

export type SegmentedOption = ChoiceOption & { isDisabled?: boolean };

type SegmentedControlProps = Omit<
  Aria.ToggleButtonGroupProps,
  'children' | 'className' | 'selectionMode'
> & {
  label: string;
  options: readonly SegmentedOption[];
};

export function SegmentedControl({
  label,
  options,
  disallowEmptySelection = true,
  ...props
}: SegmentedControlProps) {
  return (
    <Aria.ToggleButtonGroup
      {...props}
      aria-label={label}
      selectionMode="single"
      disallowEmptySelection={disallowEmptySelection}
      className="ui-segmented"
    >
      {options.map((option) => (
        <Aria.ToggleButton
          key={option.id}
          id={option.id}
          isDisabled={option.isDisabled ?? false}
          className="ui-segment"
        >
          {option.label}
        </Aria.ToggleButton>
      ))}
    </Aria.ToggleButtonGroup>
  );
}

type SelectProps = Omit<Aria.SelectProps<ChoiceOption>, 'children' | 'className'> & {
  label: string;
  options: readonly ChoiceOption[];
  description?: string;
  errorMessage?: string;
};

export function Select({ label, options, description, errorMessage, ...props }: SelectProps) {
  return (
    <Aria.Select
      {...props}
      placeholder={props.placeholder ?? 'Оберіть варіант'}
      className="ui-field"
    >
      <Aria.Label>{label}</Aria.Label>
      <Aria.Button className="ui-input ui-select-button">
        <Aria.SelectValue />
        <span aria-hidden="true">⌄</span>
      </Aria.Button>
      <FieldHelp description={description} errorMessage={errorMessage} />
      <Aria.Popover className="ui-popover">
        <Aria.ListBox items={options} className="ui-listbox">
          {(item) => (
            <Aria.ListBoxItem id={item.id} textValue={item.label} className="ui-option">
              {item.label}
            </Aria.ListBoxItem>
          )}
        </Aria.ListBox>
      </Aria.Popover>
    </Aria.Select>
  );
}

type ComboBoxProps = Omit<
  Aria.ComboBoxProps<ChoiceOption>,
  'children' | 'className' | 'items' | 'defaultItems'
> & {
  label: string;
  options: readonly ChoiceOption[];
  description?: string;
  errorMessage?: string;
  /**
   * 'server': the options are already the server's answer to the typed text, and may arrive
   * after typing started — so they are passed as controlled items and the list may open empty.
   */
  filtering?: 'client' | 'server';
  loadingLabel?: string;
  isLoading?: boolean;
};

export function ComboBox({
  label,
  options,
  description,
  errorMessage,
  loadingLabel,
  isLoading = false,
  filtering = 'client',
  ...props
}: ComboBoxProps) {
  const collection =
    filtering === 'server'
      ? { items: options, allowsEmptyCollection: true }
      : { defaultItems: options };
  return (
    <Aria.ComboBox {...props} {...collection} className="ui-field">
      <Aria.Label>{label}</Aria.Label>
      <div className="ui-combobox-input">
        <Aria.Input className="ui-input" />
        <Aria.Button className="ui-button ui-button--secondary" aria-label="Показати варіанти">
          ⌄
        </Aria.Button>
      </div>
      <FieldHelp description={description} errorMessage={errorMessage} />
      <Aria.Popover className="ui-popover">
        <Aria.ListBox<ChoiceOption>
          className="ui-listbox"
          renderEmptyState={() =>
            isLoading && loadingLabel ? loadingLabel : 'Немає відповідних варіантів'
          }
        >
          {(item) => (
            <Aria.ListBoxItem id={item.id} textValue={item.label} className="ui-option">
              {item.label}
            </Aria.ListBoxItem>
          )}
        </Aria.ListBox>
      </Aria.Popover>
    </Aria.ComboBox>
  );
}

type CheckboxProps = Omit<Aria.CheckboxProps, 'children' | 'className'> & { children: ReactNode };

export function Checkbox({ children, ...props }: CheckboxProps) {
  return (
    <Aria.Checkbox {...props} className="ui-choice">
      <span className="ui-checkbox-mark" aria-hidden="true">
        ✓
      </span>
      <span>{children}</span>
    </Aria.Checkbox>
  );
}

type SwitchProps = Omit<Aria.SwitchProps, 'children' | 'className'> & { children: ReactNode };

export function Switch({ children, ...props }: SwitchProps) {
  return (
    <Aria.Switch {...props} className="ui-choice">
      <span className="ui-switch-track" aria-hidden="true">
        <span />
      </span>
      <span>{children}</span>
    </Aria.Switch>
  );
}

type FileUploadProps = Omit<Aria.FileTriggerProps, 'children'> & {
  label: string;
  isDisabled?: boolean;
};

export function FileUpload({ label, isDisabled, ...props }: FileUploadProps) {
  return (
    <Aria.FileTrigger {...props}>
      <Button type="button" variant="secondary" isDisabled={isDisabled ?? false}>
        {label}
      </Button>
    </Aria.FileTrigger>
  );
}
