import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useAuth } from '@/app/providers/auth-provider';
import { TextField } from '@/shared/components/ui/text-field';
import { ApiError } from '@/shared/api/client';
import {
  registerSchema,
  type RegisterForm,
} from '@/features/auth/schemas/auth.schema';

export function RegisterPage() {
  const { register: registerUser } = useAuth();
  const navigate = useNavigate();
  const [formError, setFormError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterForm>({ resolver: zodResolver(registerSchema) });

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    try {
      await registerUser(values);
      navigate('/dashboard', { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 422 && err.details) {
          for (const [field, messages] of Object.entries(err.details)) {
            setError(field as keyof RegisterForm, { message: messages[0] });
          }
        } else {
          setFormError(err.message);
        }
      } else {
        setFormError('Something went wrong. Please try again.');
      }
    }
  });

  return (
    <div>
      <h2 className="fs-20 fw-bolder mb-1">Create your account</h2>
      <p className="fs-12 fw-medium text-muted mb-4">
        Your personal cashflow intelligence workspace.
      </p>
      <form onSubmit={onSubmit} className="w-100" noValidate>
        <div className="mb-3">
          <TextField
            label="Name"
            autoComplete="name"
            error={errors.name?.message}
            {...register('name')}
          />
        </div>
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
            autoComplete="new-password"
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
            {isSubmitting ? 'Creating account…' : 'Create account'}
          </button>
        </div>
      </form>
      <p className="mt-5 mb-0 text-muted fs-13 text-center">
        Already have an account?{' '}
        <Link to="/login" className="fw-bold text-primary">
          Sign in
        </Link>
      </p>
    </div>
  );
}