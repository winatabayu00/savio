import { act, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ErrorNotice } from '@/shared/components/ui/error-notice';

describe('ErrorNotice', () => {
  it('shows a safe message for frontend runtime errors', () => {
    render(<ErrorNotice />);
    act(() => window.dispatchEvent(new ErrorEvent('error', { error: new Error('internal frontend detail') })));
    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
    expect(screen.queryByText('internal frontend detail')).not.toBeInTheDocument();
  });

  it('shows a safe message for unhandled promise rejections', () => {
    render(<ErrorNotice />);
    act(() => window.dispatchEvent(new Event('unhandledrejection')));
    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
  });
});
