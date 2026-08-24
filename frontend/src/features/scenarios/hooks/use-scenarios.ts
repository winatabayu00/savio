import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addModification,
  calculateScenario,
  createScenario,
  deleteScenario,
  listScenarios,
  removeModification,
} from '@/features/scenarios/api/scenario.api';

export const scenarioKeys = { all: ['scenarios'] as const };

export function useScenarios() {
  return useQuery({ queryKey: scenarioKeys.all, queryFn: listScenarios });
}

export function useScenarioMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: scenarioKeys.all });
  const create = useMutation({ mutationFn: (name: string) => createScenario(name), onSuccess: invalidate });
  const remove = useMutation({
    mutationFn: (id: string) => deleteScenario(id),
    onSuccess: invalidate,
  });
  const addMod = useMutation({
    mutationFn: ({ id, input }: { id: string; input: { type: Parameters<typeof addModification>[1]['type']; amount: string; frequency?: string; narrative?: string } }) =>
      addModification(id, input),
    onSuccess: invalidate,
  });
  const removeMod = useMutation({
    mutationFn: ({ scenarioId, modId }: { scenarioId: string; modId: string }) => removeModification(scenarioId, modId),
    onSuccess: invalidate,
  });
  const calc = useMutation({
    mutationFn: ({ id, horizon }: { id: string; horizon?: number }) => calculateScenario(id, horizon),
    onSuccess: invalidate,
  });
  return { create, remove, addMod, removeMod, calc };
}