import { useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { Modal } from '@/shared/components/ui/modal';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { TextField } from '@/shared/components/ui/text-field';
import { useCategoryMutations, useCategories } from '@/features/categories/hooks/use-categories';
import type { CategoryType } from '@/features/categories/api/category.api';

const categorySchema = z.object({
  name: z.string().min(1, 'Category name is required'),
  type: z.enum(['INCOME', 'EXPENSE']),
});

export function CategoriesPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const { data, isLoading, isError, refetch } = useCategories();
  const { create, archive, restore } = useCategoryMutations();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<{ name: string; type: CategoryType }>({
    resolver: zodResolver(categorySchema),
    defaultValues: { name: '', type: 'EXPENSE' },
  });

  const flash = (message: string) => {
    setNotice(message);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const onSubmit = handleSubmit(async (values) => {
    try {
      await create.mutateAsync(values);
      setCreateOpen(false);
      reset();
    } catch (err) {
      const apiErr = err as ApiError;
      flash(apiErr?.message ?? 'Could not create category.');
    }
  });

  const income = data?.filter((c) => c.type === 'INCOME') ?? [];
  const expense = data?.filter((c) => c.type === 'EXPENSE') ?? [];

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Categories</h1>
          <p className="mt-1 text-sm text-gray-500">
            System categories plus your own, scoped to this workspace.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>Add Category</Button>
      </div>

      {notice ? <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div> : null}
      {isLoading ? <p className="mt-6 text-sm text-gray-500">Loading categories…</p> : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          <p>We could not load categories.</p>
          <Button variant="secondary" className="mt-2" onClick={() => void refetch()}>
            Try again
          </Button>
        </div>
      ) : null}

      <div className="mt-6 grid gap-6 md:grid-cols-2">
        {(['INCOME', 'EXPENSE'] as const).map((type) => {
          const list = type === 'INCOME' ? income : expense;
          return (
            <section key={type}>
              <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-500">
                {type}
              </h2>
              {list.length === 0 ? (
                <EmptyState title="No categories" />
              ) : (
                <ul className="mt-3 space-y-2">
                  {list.map((c) => (
                    <li
                      key={c.id}
                      className="flex items-center justify-between rounded-lg border border-gray-200 bg-white px-4 py-3"
                    >
                      <div>
                        <span className="text-sm font-medium text-gray-900">{c.name}</span>
                        {c.is_system ? (
                          <span className="ml-2 rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500">
                            system
                          </span>
                        ) : null}
                      </div>
                      {!c.is_system ? (
                        <Button
                          variant="ghost"
                          onClick={() =>
                            c.status === 'ACTIVE'
                              ? archive.mutateAsync(c.id).then(() => flash('Category archived.'))
                              : restore.mutateAsync(c.id).then(() => flash('Category restored.'))
                          }
                        >
                          {c.status === 'ACTIVE' ? 'Archive' : 'Restore'}
                        </Button>
                      ) : null}
                    </li>
                  ))}
                </ul>
              )}
            </section>
          );
        })}
      </div>

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Add category">
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <TextField label="Name" placeholder="e.g. Pets" error={errors.name?.message} {...register('name')} />
          <div>
            <label htmlFor="cat-type" className="mb-1.5 block text-sm font-medium text-gray-700">
              Type
            </label>
            <select
              id="cat-type"
              className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm"
              {...register('type')}
            >
              <option value="EXPENSE">Expense</option>
              <option value="INCOME">Income</option>
            </select>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              Create category
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}