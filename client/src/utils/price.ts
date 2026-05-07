export function formatPrice(cents: number): string {
  return (cents / 100).toFixed(2)
}

/**
 * Format a virtual sales count for display.
 * e.g. 12345 → "1.2万+" ; 500 → "500+"
 */
export function formatSales(count: number): string {
  if (count >= 10000) {
    return `${(count / 10000).toFixed(1)}万+`
  }
  return `${count}+`
}
