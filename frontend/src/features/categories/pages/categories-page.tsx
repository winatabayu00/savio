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
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Categories</h1>
          <p className="fs-13 text-muted mb-0 mt-1">
            System categories plus your own, scoped to this workspace.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>Add Category</Button>
      </div>

      {notice ? <div className="mt-4 bg-soft-primary text-primary p-3 fs-13 rounded">{notice}</div> : null}
      {isLoading ? <p className="mt-4 fs-13 text-muted">Loading categories…</p> : null}
      {isError ? (
        <div className="mt-4 alert alert-danger">
          <p>We could not load categories.</p>
          <Button variant="secondary" className="mt-2" onClick={() => void refetch()}>
            Try again
          </Button>
        </div>
      ) : null}

      <div className="row g-4 mt-4">
        {(['INCOME', 'EXPENSE'] as const).map((type) => {
          const list = type === 'INCOME' ? income : expense;
          return (
            <section key={type} className="col-12 col-md-6">
              <h2 className="fs-12 text-uppercase fw-semibold text-muted mb-3">
                {type}
              </h2>
              {list.length === 0 ? (
                <EmptyState title="No categories" />
              ) : (
                <ul className="d-flex flex-column gap-2 mt-3">
                  {list.map((c) => (
                    <li
                      key={c.id}
                      className="card card-body d-flex align-items-center justify-content-between py-3"
                    >
                      <div>
                        <span className="fs-13 fw-medium text-dark">{c.name}</span>
                        {c.is_system ? (
                          <span className="ms-2 badge bg-soft-secondary text-secondary">
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
        <form onSubmit={onSubmit} className="d-flex flex-column gap-3" noValidate>
          <TextField label="Name" placeholder="e.g. Pets" error={errors.name?.message} {...register('name')} />
          <div>
            <label htmlFor="cat-type" className="form-label">
              Type
            </label>
            <select
              id="cat-type"
              className="form-select"
              {...register('type')}
            >
              <option value="EXPENSE">Expense</option>
              <option value="INCOME">Income</option>
            </select>
          </div>
          <div className="d-flex justify-content-end gap-2 pt-2">
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