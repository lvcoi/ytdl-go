import { describe, it, expect } from 'vitest';
import { getStatusColor } from './theme';

describe('theme utils', () => {
  it('should return correct status colors', () => {
    expect(getStatusColor('completed')).toBe('text-green-400');
    expect(getStatusColor('downloading')).toBe('text-blue-400');
    expect(getStatusColor('error')).toBe('text-red-400');
    expect(getStatusColor('unknown')).toBe('text-gray-400');
  });
});
