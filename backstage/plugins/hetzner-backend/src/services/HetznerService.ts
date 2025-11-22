import {
  coreServices,
  createServiceFactory,
  createServiceRef,
  LoggerService,
} from '@backstage/backend-plugin-api';
import { Config } from '@backstage/config';
import {
  DEFAULT_NAMESPACE,
  Entity,
  stringifyEntityRef,
} from '@backstage/catalog-model';
import { Expand } from '@backstage/types';

export const HETZNER_DATA_ANNOTATION = 'hetzner.com/data';

type HetznerProjectConfig = {
  title: string;
  owner: string;
  lifecycle: string;
};

export type HetznerProjectOverview = HetznerProjectConfig & {
  totals: {
    servers: number;
    running: number;
    datacenters: string[];
  };
  generatedAt: string;
};

export type HetznerServerSummary = {
  id: number;
  name: string;
  status: string;
  datacenter: string;
  location: string;
  ipv4Address?: string | null;
  ipv6Address?: string | null;
  serverType: string;
  createdAt: string;
  locked: boolean;
  entityRef: string;
};

export type HetznerSummary = {
  project: HetznerProjectOverview;
  servers: HetznerServerSummary[];
};

type HetznerServerRecord = HetznerServerSummary & {
  entityName: string;
};

type HetznerServiceOptions = {
  client: HetznerApiClient;
  project: HetznerProjectConfig;
  locationKey: string;
  namespace: string;
};

export class HetznerService {
  readonly #client: HetznerApiClient;
  readonly #project: HetznerProjectConfig;
  readonly #locationKey: string;
  readonly #namespace: string;
  static create(options: HetznerServiceOptions) {
    return new HetznerService(options);
  }

  static fromConfig(config: Config, logger: LoggerService) {
    const backendConfig = config.getConfig('backend');
    const hetznerConfig = backendConfig.getConfig('hetzner');

    const token = hetznerConfig.getString('token');
    const configuredBaseUrl =
      hetznerConfig.getOptionalString('baseUrl') ??
      'https://api.hetzner.cloud/v1/';
    const baseUrl = configuredBaseUrl.endsWith('/')
      ? configuredBaseUrl
      : `${configuredBaseUrl}/`;
    const locationKey =
      hetznerConfig.getOptionalString('locationKey') ?? 'hetzner-cloud';
    const namespace =
      hetznerConfig.getOptionalString('namespace') ?? DEFAULT_NAMESPACE;

    const projectConfig = config.getOptionalConfig('app.hetzner.project');
    if (!projectConfig) {
      throw new Error('app.hetzner.project is required for the Hetzner plugin');
    }

    const project: HetznerProjectConfig = {
      title: projectConfig.getString('title'),
      owner: projectConfig.getString('owner'),
      lifecycle: projectConfig.getString('lifecycle'),
    };

    const client = new HetznerApiClient({
      token,
      baseUrl,
      logger,
    });

    return HetznerService.create({
      client,
      project,
      locationKey,
      namespace,
    });
  }

  private constructor(options: HetznerServiceOptions) {
    this.#client = options.client;
    this.#project = options.project;
    this.#locationKey = options.locationKey;
    this.#namespace = options.namespace;
  }

  get locationKey(): string {
    return this.#locationKey;
  }

  async snapshot(): Promise<HetznerSummary> {
    const servers = await this.listServers();
    const project = this.#buildProjectOverview(servers);
    return {
      project,
      servers,
    };
  }

  async listServers(): Promise<HetznerServerSummary[]> {
    const records = await this.#fetchServerRecords();
    return records.map(({ entityName: _entityName, ...summary }) => summary);
  }

  async getProjectOverview(): Promise<HetznerProjectOverview> {
    const servers = await this.listServers();
    return this.#buildProjectOverview(servers);
  }

  async getServerById(
    id: number,
  ): Promise<HetznerServerSummary | undefined> {
    const records = await this.#fetchServerRecords();
    const record = records.find(server => server.id === id);
    if (!record) {
      return undefined;
    }

    const { entityName: _entityName, ...summary } = record;
    return summary;
  }

  async buildResourceEntities(): Promise<Entity[]> {
    const records = await this.#fetchServerRecords();
    return records.map(record => ({
      apiVersion: 'backstage.io/v1alpha1',
      kind: 'Resource',
      metadata: {
        name: record.entityName,
        title: record.name,
        description: `Hetzner Cloud server ${record.name}`,
        annotations: {
          [HETZNER_DATA_ANNOTATION]: JSON.stringify(this.#stripInternal(record)),
        },
        tags: ['hetzner', record.datacenter.toLowerCase()],
        links: [
          {
            url: 'https://console.hetzner.cloud/',
            title: 'Open in Hetzner Cloud',
          },
        ],
      },
      spec: {
        type: 'hetzner-server',
        owner: this.#project.owner,
        lifecycle: this.#project.lifecycle,
        system: 'hetzner-cloud',
      },
    }));
  }

  #stripInternal(record: HetznerServerRecord): HetznerServerSummary {
    const { entityName: _entityName, ...summary } = record;
    return summary;
  }

  async #fetchServerRecords(): Promise<HetznerServerRecord[]> {
    const servers = await this.#client.listServers();
    return servers.map(server => {
      const entityName = this.#toEntityName(server);
      return {
        id: server.id,
        name: server.name,
        status: server.status,
        datacenter: server.datacenter?.name ?? 'unknown',
        location: server.datacenter?.location?.name ?? 'unknown',
        ipv4Address: server.public_net?.ipv4?.ip ?? null,
        ipv6Address: server.public_net?.ipv6?.ip ?? null,
        serverType: server.server_type?.name ?? 'unknown',
        createdAt: server.created,
        locked: server.locked ?? false,
        entityRef: stringifyEntityRef({
          kind: 'Resource',
          namespace: this.#namespace,
          name: entityName,
        }),
        entityName,
      };
    });
  }

  #toEntityName(server: HetznerServerApi): string {
    const base = `${server.name}-${server.id}`
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '');
    return base || `hetzner-${server.id}`;
  }

  #buildProjectOverview(
    servers: HetznerServerSummary[],
  ): HetznerProjectOverview {
    const datacenters = Array.from(
      new Set(servers.map(server => server.datacenter)),
    );
    const running = servers.filter(server => server.status === 'running');

    return {
      ...this.#project,
      totals: {
        servers: servers.length,
        running: running.length,
        datacenters,
      },
      generatedAt: new Date().toISOString(),
    };
  }
}

type HetznerApiClientOptions = {
  token: string;
  baseUrl: string;
  logger: LoggerService;
};

type HetznerPagination = {
  next_page?: number | null;
};

type HetznerServerListResponse = {
  servers: HetznerServerApi[];
  meta?: {
    pagination?: HetznerPagination;
  };
};

type HetznerServerApi = {
  id: number;
  name: string;
  status: string;
  locked?: boolean;
  created: string;
  datacenter?: {
    name: string;
    location?: {
      name: string;
    };
  };
  public_net?: {
    ipv4?: { ip?: string };
    ipv6?: { ip?: string };
  };
  server_type?: {
    name: string;
  };
};

class HetznerApiClient {
  readonly #token: string;
  readonly #baseUrl: string;
  readonly #logger: LoggerService;

  constructor(options: HetznerApiClientOptions) {
    this.#token = options.token;
    this.#baseUrl = options.baseUrl;
    this.#logger = options.logger;
  }

  async listServers(): Promise<HetznerServerApi[]> {
    return this.#paginate<HetznerServerListResponse, HetznerServerApi>(
      'servers',
      response => response.servers,
    );
  }

  async #paginate<TResponse, TItem>(
    path: string,
    selector: (response: TResponse) => TItem[],
  ): Promise<TItem[]> {
    const results = new Array<TItem>();
    let page = 1;
    let hasNext = true;

    while (hasNext) {
      const url = new URL(path, this.#baseUrl);
      url.searchParams.set('page', String(page));
      url.searchParams.set('per_page', '50');

      const response = await this.#request<TResponse>(url);
      results.push(...selector(response));

      const pagination = (response as HetznerServerListResponse).meta
        ?.pagination;
      if (!pagination?.next_page) {
        hasNext = false;
      } else {
        page = pagination.next_page;
      }
    }

    return results;
  }

  async #request<TResponse>(url: URL): Promise<TResponse> {
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${this.#token}`,
        Accept: 'application/json',
      },
    });

    if (!response.ok) {
      const body = await response.text();
      this.#logger.error('Hetzner API request failed', {
        status: response.status,
        url: url.toString(),
        body,
      });
      throw new Error(
        `Hetzner API returned ${response.status} for ${url.toString()}`,
      );
    }

    return (await response.json()) as TResponse;
  }
}

export const hetznerServiceRef = createServiceRef<Expand<HetznerService>>({
  id: 'hetzner.service',
  defaultFactory: async service =>
    createServiceFactory({
      service,
      deps: {
        config: coreServices.rootConfig,
        logger: coreServices.logger,
      },
      async factory({ config, logger }) {
        return HetznerService.fromConfig(config, logger);
      },
    }),
});

export type HetznerServiceApi = typeof hetznerServiceRef.T;
