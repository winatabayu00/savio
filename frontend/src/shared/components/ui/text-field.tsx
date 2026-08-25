import {
  forwardRef,
  type InputHTMLAttributes,
} from 'react';

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

export const TextField = forwardRef<HTMLInputElement, TextFieldProps>(
  function TextField({ label, error, id, ...props }, ref) {
    const inputId = id ?? props.name;
    return (
      <div>
        <label htmlFor={inputId} className="form-label">
          {label}
        </label>
        <input
          ref={ref}
          id={inputId}
          aria-invalid={error ? true : undefined}
          className={`form-control ${error ? 'is-invalid' : ''}`}
          {...props}
        />
        {error ? (
          <p role="alert" className="invalid-feedback d-block mt-1 fs-12">
            {error}
          </p>
        ) : null}
      </div>
    );
  },
);