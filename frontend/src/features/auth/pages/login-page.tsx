import { useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { TextField } from '@/shared/components/ui/text-field';
import { ApiError, errorMessage } from '@/shared/api/client';
import { loginSchema, type LoginForm } from '@/features/auth/schemas/auth.schema';

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [formError, setFormError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) });

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    try {
      await login(values);
      const from = (location.state as { from?: string } | null)?.from;
      navigate(from || '/dashboard', { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 422 && err.details) {
          for (const [field, messages] of Object.entries(err.details)) {
            setError(field as keyof LoginForm, { message: messages[0] });
          }
        } else {
          setFormError(err.message);
        }
      } else {
        setFormError(errorMessage(err));
      }
    }
  });

  return (
    <div>
      <h1 className="text-xl font-semibold">Sign in to Savio</h1>
      <p className="mt-1 text-sm text-gray-500">
        Understand your cashflow, forecast your future.
      </p>
      <form onSubmit={onSubmit} className="mt-6 space-y-4" noValidate>
        <TextField
          label="Email"
          type="email"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />
        <TextField
          label="Password"
          type="password"
          autoComplete="current-password"
          error={errors.password?.message}
          {...register('password')}
        />
        {formError ? (
          <p role="alert" className="text-sm text-red-600">
            {formError}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full rounded-lg bg-brand px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-light disabled:opacity-60"
        >
          {isSubmitting ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
      <p className="mt-4 text-center text-sm text-gray-500">
        No account?{' '}
        <a href="/register" className="font-medium text-brand hover:underline">
          Create one
        </a>
      </p>
    </div>
  );
}
