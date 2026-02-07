import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption } from './table';

describe('Table components', () => {
  describe('Table', () => {
    it('renders table element', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Content</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByRole('table')).toBeInTheDocument();
    });

    it('applies base styles', () => {
      render(
        <Table data-testid="table">
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('table')).toHaveClass('w-full', 'text-sm');
    });

    it('accepts custom className', () => {
      render(
        <Table className="custom-table" data-testid="table">
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('table')).toHaveClass('custom-table');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table ref={ref}>
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableElement);
    });
  });

  describe('TableHeader', () => {
    it('renders thead element', () => {
      render(
        <Table>
          <TableHeader data-testid="header">
            <TableRow><TableHead>Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(screen.getByTestId('header').tagName).toBe('THEAD');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableHeader className="custom-header" data-testid="header">
            <TableRow><TableHead>Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(screen.getByTestId('header')).toHaveClass('custom-header');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableHeader ref={ref}>
            <TableRow><TableHead>Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableSectionElement);
    });
  });

  describe('TableBody', () => {
    it('renders tbody element', () => {
      render(
        <Table>
          <TableBody data-testid="body">
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('body').tagName).toBe('TBODY');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableBody className="custom-body" data-testid="body">
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('body')).toHaveClass('custom-body');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableBody ref={ref}>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableSectionElement);
    });
  });

  describe('TableFooter', () => {
    it('renders tfoot element', () => {
      render(
        <Table>
          <TableFooter data-testid="footer">
            <TableRow><TableCell>Footer</TableCell></TableRow>
          </TableFooter>
        </Table>
      );
      expect(screen.getByTestId('footer').tagName).toBe('TFOOT');
    });

    it('applies footer styles', () => {
      render(
        <Table>
          <TableFooter data-testid="footer">
            <TableRow><TableCell>Footer</TableCell></TableRow>
          </TableFooter>
        </Table>
      );
      expect(screen.getByTestId('footer')).toHaveClass('border-t', 'font-medium');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableFooter className="custom-footer" data-testid="footer">
            <TableRow><TableCell>Footer</TableCell></TableRow>
          </TableFooter>
        </Table>
      );
      expect(screen.getByTestId('footer')).toHaveClass('custom-footer');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableFooter ref={ref}>
            <TableRow><TableCell>Footer</TableCell></TableRow>
          </TableFooter>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableSectionElement);
    });
  });

  describe('TableRow', () => {
    it('renders tr element', () => {
      render(
        <Table>
          <TableBody>
            <TableRow data-testid="row"><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('row').tagName).toBe('TR');
    });

    it('applies row styles', () => {
      render(
        <Table>
          <TableBody>
            <TableRow data-testid="row"><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('row')).toHaveClass('border-b', 'transition-colors');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow className="custom-row" data-testid="row"><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('row')).toHaveClass('custom-row');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableBody>
            <TableRow ref={ref}><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableRowElement);
    });
  });

  describe('TableHead', () => {
    it('renders th element', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow><TableHead data-testid="head">Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(screen.getByTestId('head').tagName).toBe('TH');
    });

    it('applies head styles', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow><TableHead data-testid="head">Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(screen.getByTestId('head')).toHaveClass('h-12', 'px-4', 'font-medium');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow><TableHead className="custom-head" data-testid="head">Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(screen.getByTestId('head')).toHaveClass('custom-head');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableHeader>
            <TableRow><TableHead ref={ref}>Header</TableHead></TableRow>
          </TableHeader>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableCellElement);
    });
  });

  describe('TableCell', () => {
    it('renders td element', () => {
      render(
        <Table>
          <TableBody>
            <TableRow><TableCell data-testid="cell">Cell content</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('cell').tagName).toBe('TD');
    });

    it('applies cell styles', () => {
      render(
        <Table>
          <TableBody>
            <TableRow><TableCell data-testid="cell">Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('cell')).toHaveClass('p-4', 'align-middle');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow><TableCell className="custom-cell" data-testid="cell">Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('cell')).toHaveClass('custom-cell');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableBody>
            <TableRow><TableCell ref={ref}>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableCellElement);
    });
  });

  describe('TableCaption', () => {
    it('renders caption element', () => {
      render(
        <Table>
          <TableCaption data-testid="caption">Table caption</TableCaption>
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('caption').tagName).toBe('CAPTION');
    });

    it('applies caption styles', () => {
      render(
        <Table>
          <TableCaption data-testid="caption">Caption</TableCaption>
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('caption')).toHaveClass('mt-4', 'text-sm');
    });

    it('accepts custom className', () => {
      render(
        <Table>
          <TableCaption className="custom-caption" data-testid="caption">Caption</TableCaption>
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByTestId('caption')).toHaveClass('custom-caption');
    });

    it('forwards ref', () => {
      const ref = { current: null };
      render(
        <Table>
          <TableCaption ref={ref}>Caption</TableCaption>
          <TableBody>
            <TableRow><TableCell>Cell</TableCell></TableRow>
          </TableBody>
        </Table>
      );
      expect(ref.current).toBeInstanceOf(HTMLTableCaptionElement);
    });
  });

  describe('Full table composition', () => {
    it('renders complete table with all components', () => {
      render(
        <Table>
          <TableCaption>Employee List</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>John Doe</TableCell>
              <TableCell>john@example.com</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter>
            <TableRow>
              <TableCell colSpan={2}>Total: 1 employee</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      );

      expect(screen.getByText('Employee List')).toBeInTheDocument();
      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Email')).toBeInTheDocument();
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('john@example.com')).toBeInTheDocument();
      expect(screen.getByText('Total: 1 employee')).toBeInTheDocument();
    });
  });
});
