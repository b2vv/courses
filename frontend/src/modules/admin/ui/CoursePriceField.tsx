import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { parseUahAmount } from '../../../contracts/course';
import type { CoursePrice, CourseTree } from '../../../contracts/course';
import { formatPrice } from '../../../presentation/format-price';
import { Button, TextField } from '../../../presentation/ui';

type CoursePriceFieldProps = {
  readonly course: CourseTree;
  /** True while the course form has unsaved edits: a price save would reset them. */
  readonly hasUnsavedChanges: boolean;
  readonly isBusy: boolean;
  readonly onSetPrice: (current: CourseTree, price: CoursePrice | null) => Promise<void>;
};

function amountText(price: CoursePrice | null): string {
  if (price === null) return '';
  const whole = Math.floor(price.minor / 100);
  const fraction = price.minor % 100;
  return fraction === 0 ? String(whole) : `${whole}.${String(fraction).padStart(2, '0')}`;
}

/**
 * The price is saved apart from the course form, over the stored course. The server refuses a
 * first price together with any other change (it keeps access for learners already in the
 * course), so the author saves the course first and then the price — this control says so
 * instead of letting a combined save fail.
 */
export function CoursePriceField({
  course,
  hasUnsavedChanges,
  isBusy,
  onSetPrice,
}: CoursePriceFieldProps) {
  const { t, i18n } = useTranslation();
  const [text, setText] = useState(() => amountText(course.price));
  const [touched, setTouched] = useState(false);
  const [pending, setPending] = useState(false);
  const parsed = parseUahAmount(text);
  const unchanged = parsed !== null && parsed.minor === course.price?.minor;
  const disabled = hasUnsavedChanges || isBusy || pending;

  async function submit(price: CoursePrice | null) {
    setPending(true);
    try {
      await onSetPrice(course, price);
      setText(amountText(price));
      setTouched(false);
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="admin-lifecycle admin-price ui-card" aria-labelledby="course-price-heading">
      <div>
        <p className="eyebrow">{t('administration.price.eyebrow')}</p>
        <h2 id="course-price-heading">{t('administration.price.heading')}</h2>
        <p className="admin-price-current">
          {course.price === null
            ? t('administration.price.free')
            : t('administration.price.current', {
                amount: formatPrice(course.price, i18n.resolvedLanguage),
              })}
        </p>
      </div>
      {course.price === null && (
        <p className="admin-price-note">{t('administration.price.firstPriceNote')}</p>
      )}
      {hasUnsavedChanges && (
        <p className="admin-price-note">{t('administration.price.saveCourseFirst')}</p>
      )}
      <form
        className="admin-row-actions"
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          setTouched(true);
          if (parsed !== null && !unchanged) void submit(parsed);
        }}
      >
        <TextField
          label={t('administration.price.amount')}
          value={text}
          onChange={setText}
          inputMode="decimal"
          isDisabled={disabled}
          isInvalid={touched && parsed === null}
          {...(touched && parsed === null
            ? { errorMessage: t('administration.price.invalid') }
            : { description: t('administration.price.amountHint') })}
        />
        <Button type="submit" isDisabled={disabled || unchanged}>
          {pending ? t('administration.saving') : t('administration.price.save')}
        </Button>
        {course.price !== null && (
          <Button
            type="button"
            variant="secondary"
            isDisabled={disabled}
            onPress={() => void submit(null)}
          >
            {t('administration.price.makeFree')}
          </Button>
        )}
      </form>
    </section>
  );
}
