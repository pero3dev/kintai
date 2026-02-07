import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from './card';

describe('Card components', () => {
  describe('Card', () => {
    it('renders children', () => {
      render(<Card>Card content</Card>);
      expect(screen.getByText('Card content')).toBeInTheDocument();
    });

    it('applies base styles', () => {
      render(<Card data-testid="card">Content</Card>);
      const card = screen.getByTestId('card');
      expect(card).toHaveClass('rounded-lg', 'border', 'bg-card', 'shadow-sm');
    });

    it('accepts custom className', () => {
      render(<Card className="custom-card" data-testid="card">Content</Card>);
      expect(screen.getByTestId('card')).toHaveClass('custom-card');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(<Card ref={ref}>Content</Card>);
      expect(ref.current).toBeInstanceOf(HTMLDivElement);
    });
  });

  describe('CardHeader', () => {
    it('renders children', () => {
      render(<CardHeader>Header content</CardHeader>);
      expect(screen.getByText('Header content')).toBeInTheDocument();
    });

    it('applies header styles', () => {
      render(<CardHeader data-testid="header">Header</CardHeader>);
      expect(screen.getByTestId('header')).toHaveClass('flex', 'flex-col', 'p-6');
    });

    it('accepts custom className', () => {
      render(<CardHeader className="custom-header" data-testid="header">Header</CardHeader>);
      expect(screen.getByTestId('header')).toHaveClass('custom-header');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(<CardHeader ref={ref}>Header</CardHeader>);
      expect(ref.current).toBeInstanceOf(HTMLDivElement);
    });
  });

  describe('CardTitle', () => {
    it('renders as h3', () => {
      render(<CardTitle>Title</CardTitle>);
      expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('Title');
    });

    it('applies title styles', () => {
      render(<CardTitle data-testid="title">Title</CardTitle>);
      expect(screen.getByTestId('title')).toHaveClass('text-2xl', 'font-semibold');
    });

    it('accepts custom className', () => {
      render(<CardTitle className="custom-title" data-testid="title">Title</CardTitle>);
      expect(screen.getByTestId('title')).toHaveClass('custom-title');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(<CardTitle ref={ref}>Title</CardTitle>);
      expect(ref.current).toBeInstanceOf(HTMLHeadingElement);
    });
  });

  describe('CardDescription', () => {
    it('renders as paragraph', () => {
      render(<CardDescription>Description</CardDescription>);
      expect(screen.getByText('Description').tagName).toBe('P');
    });

    it('applies description styles', () => {
      render(<CardDescription data-testid="desc">Description</CardDescription>);
      expect(screen.getByTestId('desc')).toHaveClass('text-sm', 'text-muted-foreground');
    });

    it('accepts custom className', () => {
      render(<CardDescription className="custom-desc" data-testid="desc">Description</CardDescription>);
      expect(screen.getByTestId('desc')).toHaveClass('custom-desc');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(<CardDescription ref={ref}>Description</CardDescription>);
      expect(ref.current).toBeInstanceOf(HTMLParagraphElement);
    });
  });

  describe('CardContent', () => {
    it('renders children', () => {
      render(<CardContent>Main content</CardContent>);
      expect(screen.getByText('Main content')).toBeInTheDocument();
    });

    it('applies content styles', () => {
      render(<CardContent data-testid="content">Content</CardContent>);
      expect(screen.getByTestId('content')).toHaveClass('p-6', 'pt-0');
    });

    it('accepts custom className', () => {
      render(<CardContent className="custom-content" data-testid="content">Content</CardContent>);
      expect(screen.getByTestId('content')).toHaveClass('custom-content');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(<CardContent ref={ref}>Content</CardContent>);
      expect(ref.current).toBeInstanceOf(HTMLDivElement);
    });
  });

  describe('CardFooter', () => {
    it('renders children', () => {
      render(<CardFooter>Footer content</CardFooter>);
      expect(screen.getByText('Footer content')).toBeInTheDocument();
    });

    it('applies footer styles', () => {
      render(<CardFooter data-testid="footer">Footer</CardFooter>);
      expect(screen.getByTestId('footer')).toHaveClass('flex', 'items-center', 'p-6');
    });

    it('accepts custom className', () => {
      render(<CardFooter className="custom-footer" data-testid="footer">Footer</CardFooter>);
      expect(screen.getByTestId('footer')).toHaveClass('custom-footer');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(<CardFooter ref={ref}>Footer</CardFooter>);
      expect(ref.current).toBeInstanceOf(HTMLDivElement);
    });
  });

  describe('Full card composition', () => {
    it('renders complete card with all components', () => {
      render(
        <Card data-testid="full-card">
          <CardHeader>
            <CardTitle>Test Card</CardTitle>
            <CardDescription>This is a test description</CardDescription>
          </CardHeader>
          <CardContent>
            <p>Card body content</p>
          </CardContent>
          <CardFooter>
            <button>Action</button>
          </CardFooter>
        </Card>
      );

      expect(screen.getByRole('heading', { name: 'Test Card' })).toBeInTheDocument();
      expect(screen.getByText('This is a test description')).toBeInTheDocument();
      expect(screen.getByText('Card body content')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Action' })).toBeInTheDocument();
    });
  });
});
