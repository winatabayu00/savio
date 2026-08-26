import { useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { useAuth } from '@/app/providers/auth-provider';
import type { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { TextField } from '@/shared/components/ui/text-field';
import { useAccountMutations } from '@/features/accounts/hooks/use-accounts';
import type { Account, AccountType } from '@/features/accounts/types/account.types';

const accountSchema = z.object({
  name: z.string().min(1, 'Account name is required'),
  type: z.string().min(1, 'Account type is required'),
  opening_balance: z
    .string()
    .refine((v) => v.trim() !== '' && !Number.isNaN(Number(v)) && Number(v) >= 0, 'Opening balance must be zero or positive'),
});

const ACCOUNT_TYPES: AccountType[] = ['CASH', 'BANK', 'EWALLET', 'SAVINGS', 'OTHER'];

interface AccountFormProps {
  account?: Account;
  onDone: () => void;
}

export function AccountForm({ account, onDone }: AccountFormProps) {
  const { auth } = useAuth();
  const { create, update } = useAccountMutations();
  const [formError, setFormError] = useState<string | null>(null);
  const [amount, setAmount] = useState(() =>
    account ? String(account.opening_balance / 100) : '',
  );
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm({
    resolver: zodResolver(accountSchema),
    defaultValues: account
      ? { name: account.name, type: account.type, opening_balance: String(account.opening_balance / 100) }
      : { name: '', type: 'BANK', opening_balance: '' },
  });

  const submitted = handleSubmit(async (values) => {
    setFormError(null);
    const minor = Math.round(Number(amount) * 100);
    try {
      if (account) {
        await update.mutateAsync({
          id: account.id,
          input: {
            name: values.name,
            type: values.type as AccountType,
            institution_name: undefined,
            description: undefined,
            version: account.version,
          },
        });
      } else {
        await create.mutateAsync({
          name: values.name,
          type: values.type as AccountType,
          currency: auth?.workspace.base_currency ?? 'IDR',
          opening_balance: minor,
        });
      }
      onDone();
    } catch (err) {
      const apiErr = err as ApiError;
      setFormError(apiErr?.message ?? 'Something went wrong. Please try again.');
    }
  });

  return (
    <form onSubmit={submitted} noValidate>
      <div className="mb-3">
        <TextField
          label="Account name"
          placeholder="e.g. Daily Wallet"
          error={errors.name?.message as string | undefined}
          {...register('name')}
        />
      </div>
      <div className="mb-3">
        <label htmlFor="type" className="form-label">
          Account type
        </label>
        <select id="type" className="form-select" {...register('type')}>
          {ACCOUNT_TYPES.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </div>
      {!account ? (
        <div className="mb-3">
          <label htmlFor="amount" className="form-label">
            Opening balance ({auth?.workspace.base_currency ?? 'IDR'})
          </label>
          <input
            id="amount"
            inputMode="decimal"
            className="form-control"
            value={amount}
            onChange={(e) => {
              setAmount(e.target.value);
              setValue('opening_balance', e.target.value, { shouldValidate: true });
            }}
          />
          {errors.opening_balance ? (
            <p role="alert" className="text-danger fs-12 mt-1">
              {errors.opening_balance.message as string}
            </p>
          ) : null}
        </div>
      ) : null}
      {formError ? (
        <p role="alert" className="text-danger fs-13 mb-3">
          {formError}
        </p>
      ) : null}
      <div className="d-flex justify-content-end gap-2 pt-2">
        <Button type="button" variant="secondary" onClick={onDone}>
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {account ? 'Save changes' : 'Create account'}
        </Button>
      </div>
    </form>
  );
}