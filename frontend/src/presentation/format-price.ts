export type DisplayPrice = { readonly minor: number; readonly currency: 'UAH' };

/** One price format for the storefront and the learning page, so they cannot drift. */
export function formatPrice(price: DisplayPrice, language: string | undefined): string {
  return new Intl.NumberFormat(language === 'en' ? 'en-US' : 'uk-UA', {
    style: 'currency',
    currency: price.currency,
    currencyDisplay: 'narrowSymbol',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(price.minor / 100);
}
