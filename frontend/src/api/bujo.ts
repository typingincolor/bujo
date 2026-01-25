import { Entry } from '@/types/bujo';

export async function getEntriesForDateRange(
  startDate: string,
  endDate: string
): Promise<Entry[]> {
  const response = await fetch(
    `/api/entries?start=${startDate}&end=${endDate}`
  );
  if (!response.ok) {
    throw new Error('Failed to fetch entries');
  }
  return response.json();
}
