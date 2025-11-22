export const HETZNER_DATA_ANNOTATION = 'hetzner.com/data';

export type HetznerProjectOverview = {
  title: string;
  owner: string;
  lifecycle: string;
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

export type HetznerSummaryResponse = {
  project: HetznerProjectOverview;
  servers: HetznerServerSummary[];
};
