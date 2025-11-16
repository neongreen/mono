import { Entity } from '@backstage/catalog-model';

export const isHetznerResource = (entity: Entity): boolean => {
  const hetznerData = entity.metadata.annotations?.['hetzner.com/data'];
  try {
    return hetznerData !== undefined && JSON.parse(hetznerData) !== null;
  } catch (e) {
    return false;
  }
};
