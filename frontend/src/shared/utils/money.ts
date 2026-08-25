// ponytail: register/login omit workspace currency -> "", normalize to IDR
function normalize(currency?: string): string {
  return (currency && currency.trim().toUpperCase()) || 'IDR';
}

function formatter(currency: string, fractionDigits: number): Intl.NumberFormat {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
}

// Minor units (integer) to a decimal-safe display string.
export function formatMoney(amount: number, currency = 'IDR'): string {
  const cur = normalize(currency);
  if (!Number.isFinite(amount)) return '—';
  try {
    return formatter(cur, cur === 'IDR' ? 0 : 2).format(amount / 100);
  } catch {
    return formatter('IDR', 0).format(amount / 100);
  }
}

// Backend amounts arrive as major-unit decimal strings (e.g. "1500000.00");
// format them directly without a scale shift.
export function formatAmountString(amount: string, currency = 'IDR'): string {
  const cur = normalize(currency);
  const value = Number(amount);
  if (!Number.isFinite(value)) return '—';
  try {
    return formatter(cur, cur === 'IDR' ? 0 : 2).format(value);
  } catch {
    return formatter('IDR', 0).format(value);
  }
}
