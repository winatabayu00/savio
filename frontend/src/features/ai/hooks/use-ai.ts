import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  askCopilot,
  getInsight,
  aiStatus,
  listConversations,
  getConversation,
  createConversation,
  sendConversationMessage,
  deleteConversation,
} from '@/features/ai/api/ai.api';

export const aiKeys = {
  all: ['ai'] as const,
  conversations: ['ai', 'conversations'] as const,
  conversation: (id: string) => ['ai', 'conversations', id] as const,
};

export function useAIStatus() {
  return useQuery({ queryKey: [...aiKeys.all, 'status'], queryFn: aiStatus, retry: 1, staleTime: 60_000 });
}

export function useInsight() {
  return useQuery({ queryKey: [...aiKeys.all, 'insight'], queryFn: getInsight, retry: 1, staleTime: 60_000 });
}

export function useCopilot() {
  const qc = useQueryClient();
  const ask = useMutation({ mutationFn: (question: string) => askCopilot(question) });
  return { ask, clear: () => qc.removeQueries({ queryKey: aiKeys.all }) };
}

export function useConversations() {
  return useQuery({ queryKey: aiKeys.conversations, queryFn: listConversations });
}

export function useConversation(id: string | null) {
  return useQuery({
    queryKey: aiKeys.conversation(id ?? ''),
    queryFn: () => getConversation(id!),
    enabled: Boolean(id),
  });
}

export function useConversationMutations() {
  const qc = useQueryClient();
  const create = useMutation({
    mutationFn: createConversation,
    onSuccess: (row) => {
      qc.setQueryData(aiKeys.conversation(row.id), row);
      void qc.invalidateQueries({ queryKey: aiKeys.conversations });
    },
  });
  const send = useMutation({
    mutationFn: ({ id, question }: { id: string; question: string }) => sendConversationMessage(id, question),
    onSuccess: (row) => {
      qc.setQueryData(aiKeys.conversation(row.id), row);
      void qc.invalidateQueries({ queryKey: aiKeys.conversations });
    },
  });
  const remove = useMutation({
    mutationFn: deleteConversation,
    onSuccess: (_, id) => {
      qc.removeQueries({ queryKey: aiKeys.conversation(id) });
      void qc.invalidateQueries({ queryKey: aiKeys.conversations });
    },
  });
  return { create, send, remove };
}
