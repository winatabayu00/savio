// ponytail: register/login omit workspace currency -> "", normalize to IDR
function normalize(currency?: string): string {
  return (currency && currency.trim()) || 'IDR';
}

// Minor units (integer) to a decimal-safe display string.
export function formatMoney(amount: number, currency = 'IDR'): string {
  const cur = normalize(currency);
  const opts: Intl.NumberFormatOptions = { style: 'currency', currency: cur };
  if (cur === 'IDR') {
    opts.minimumFractionDigits = 0;
    opts.maximumFractionDigits = 0;
  } else {
    opts.minimumFractionDigits = 2;
    opts.maximumFractionDigits = 2;
  }
  return new Intl.NumberFormat('id-ID', opts).format(amount / 100);
}

// Backend amounts arrive as major-unit decimal strings (e.g. "1500000.00");
// format them directly without a scale shift.
export function formatAmountString(amount: string, currency = 'IDR'): string {
  const cur = normalize(currency);
  const opts: Intl.NumberFormatOptions = { style: 'currency', currency: cur };
  if (cur === 'IDR') {
    opts.maximumFractionDigits = 0;
  } else {
    opts.minimumFractionDigits = 2;
    opts.maximumFractionDigits = 2;
  }
  return new Intl.NumberFormat('id-ID', opts).format(Number(amount));
}