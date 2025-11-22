import { hetznerPlugin } from './plugin';

describe('hetzner', () => {
  it('should export plugin', () => {
    expect(hetznerPlugin).toBeDefined();
  });
});
