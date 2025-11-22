import { LoggerService } from '@backstage/backend-plugin-api';
import {
  EntityProvider,
  EntityProviderConnection,
} from '@backstage/plugin-catalog-node';
import type { HetznerServiceApi } from '../services/HetznerService';

export class HetznerEntityProvider implements EntityProvider {
  #connection?: EntityProviderConnection;
  readonly #logger: LoggerService;
  readonly #service: HetznerServiceApi;

  constructor(options: { logger: LoggerService; service: HetznerServiceApi }) {
    this.#logger = options.logger;
    this.#service = options.service;
  }

  getProviderName(): string {
    return 'hetzner-cloud';
  }

  async connect(connection: EntityProviderConnection): Promise<void> {
    this.#connection = connection;
  }

  async run(): Promise<void> {
    if (!this.#connection) {
      throw new Error('HetznerEntityProvider is not initialized');
    }

    const entities = await this.#service.buildResourceEntities();
    await this.#connection.applyMutation({
      type: 'full',
      entities: entities.map(entity => ({
        entity,
        locationKey: this.#service.locationKey,
      })),
    });

    this.#logger.info('hetzner-cloud: refreshed entities', {
      count: entities.length,
    });
  }
}
