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
      <h2 className="fs-20 fw-bolder mb-1">Login</h2>
      <p className="fs-12 fw-medium text-muted mb-4">
        Understand your cashflow, forecast your future.
      </p>
      <form onSubmit={onSubmit} className="w-100" noValidate>
        <div className="mb-3">
          <TextField
            label="Email"
            type="email"
            autoComplete="email"
            error={errors.email?.message}
            {...register('email')}
          />
        </div>
        <div className="mb-3">
          <TextField
            label="Password"
            type="password"
            autoComplete="current-password"
            error={errors.password?.message}
            {...register('password')}
          />
        </div>
        {formError ? (
          <p role="alert" className="text-danger fs-13">
            {formError}
          </p>
        ) : null}
        <div className="mt-4">
          <button
            type="submit"
            disabled={isSubmitting}
            className="btn btn-lg btn-primary w-100"
          >
            {isSubmitting ? 'Signing in…' : 'Sign in'}
          </button>
        </div>
      </form>
      <p className="mt-5 mb-0 text-muted fs-13 text-center">
        No account?{' '}
        <a href="/register" className="fw-bold text-primary">
          Create one
        </a>
      </p>
    </div>
  );
}
