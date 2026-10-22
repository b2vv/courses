import { useTranslation } from 'react-i18next';
import type { MembershipCard } from '../../modules/enrollment';
import { EmptyState, Link, ProgressBar } from '../ui';

export function MyLearningList({ items }: { readonly items: readonly MembershipCard[] }) {
  const { t } = useTranslation();
  if (items.length === 0)
    return (
      <EmptyState title={t('enrollment.myLearning.title')}>
        <p>{t('enrollment.myLearning.empty')}</p>
        <Link href="/courses">{t('enrollment.myLearning.browse')}</Link>
      </EmptyState>
    );
  return (
    <section aria-labelledby="my-learning-title">
      <h1 id="my-learning-title">{t('enrollment.myLearning.title')}</h1>
      <ul className="my-learning-list">
        {items.map((item) => (
          <li key={item.courseId} className="my-learning-card">
            <h2>{item.title}</h2>
            <p className={`ui-pill ui-pill--${item.state === 'blocked' ? 'warning' : 'success'}`}>
              {item.state === 'blocked' ? t('enrollment.blocked') : t('enrollment.member')}
            </p>
            {item.state === 'blocked' && <output>{t('enrollment.blockedNote')}</output>}
            <ProgressBar
              label={t('enrollment.progressLabel')}
              value={item.progressPercent}
              minValue={0}
              maxValue={100}
              valueLabel={t('enrollment.progress', { percent: item.progressPercent })}
            />
            <Link href={`/learn/${item.slug}`}>{t('enrollment.startLearning')}</Link>
          </li>
        ))}
      </ul>
    </section>
  );
}
