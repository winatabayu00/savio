import {
  forwardRef,
  type ButtonHTMLAttributes,
  type ReactNode,
} from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  children: ReactNode;
}

const VARIANTS: Record<NonNullable<ButtonProps['variant']>, string> = {
  primary: 'btn-primary',
  secondary: 'btn-light-brand',
  ghost: 'btn-link',
  danger: 'btn-danger',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  function Button({ variant = 'primary', className = '', children, ...props }, ref) {
    return (
      <button
        ref={ref}
        className={`btn ${VARIANTS[variant]} ${className}`}
        {...props}
      >
        {children}
      </button>
    );
  },
);