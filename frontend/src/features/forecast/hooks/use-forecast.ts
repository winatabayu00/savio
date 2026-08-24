import { useQuery } from '@tanstack/react-query';
import { getForecast } from '@/features/forecast/api/forecast.api';

export const forecastKeys = { all: ['forecast'] as const };

export function useForecast(horizon: number) {
  return useQuery({
    queryKey: [...forecastKeys.all, horizon],
    queryFn: () => getForecast(horizon),
    staleTime: 60_000,
  });
}