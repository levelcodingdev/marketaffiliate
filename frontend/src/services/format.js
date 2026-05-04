export function money(cents = 0) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format((Number(cents) || 0) / 100);
}

export function percent(value = 0) {
  return `${Number(value || 0).toFixed(0)}%`;
}
