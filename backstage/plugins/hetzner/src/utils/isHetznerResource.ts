import { Entity } from '@backstage/catalog-model';
import { HETZNER_DATA_ANNOTATION } from '../types';

export const isHetznerResource = (entity: Entity): boolean => {
  const annotation = entity.metadata.annotations?.[HETZNER_DATA_ANNOTATION];
  if (!annotation) {
    return false;
  }

  try {
    const parsed = JSON.parse(annotation);
    return parsed !== null && typeof parsed === 'object';
  } catch {
    return false;
  }
};
