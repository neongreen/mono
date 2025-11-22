import { Chip, Grid, Typography } from '@material-ui/core';
import {
  Content,
  ContentHeader,
  Header,
  HeaderLabel,
  InfoCard,
  Page,
  Progress,
  ResponseErrorPanel,
  SupportButton,
  Table,
  TableColumn,
} from '@backstage/core-components';
import {
  discoveryApiRef,
  fetchApiRef,
  identityApiRef,
  useApi,
} from '@backstage/core-plugin-api';
import { EntityRefLink } from '@backstage/plugin-catalog-react';
import { ResponseError } from '@backstage/errors';
import useAsync from 'react-use/lib/useAsync';
import {
  HetznerServerSummary,
  HetznerSummaryResponse,
} from '../../types';

const columns: TableColumn<HetznerServerSummary>[] = [
  {
    title: 'Name',
    field: 'name',
    highlight: true,
    render: row => (
      <EntityRefLink entityRef={row.entityRef} title={row.name} />
    ),
  },
  { title: 'Status', field: 'status' },
  { title: 'Server type', field: 'serverType' },
  { title: 'Datacenter', field: 'datacenter' },
  {
    title: 'IPv4',
    field: 'ipv4Address',
    render: row => row.ipv4Address ?? '—',
  },
  {
    title: 'IPv6',
    field: 'ipv6Address',
    render: row => row.ipv6Address ?? '—',
  },
  {
    title: 'Created',
    field: 'createdAt',
    render: row =>
      new Date(row.createdAt).toLocaleString(undefined, {
        dateStyle: 'medium',
        timeStyle: 'short',
      }),
  },
];

export const HetznerPage = () => {
  const fetchApi = useApi(fetchApiRef);
  const identityApi = useApi(identityApiRef);
  const discoveryApi = useApi(discoveryApiRef);

  const { value: identity } = useAsync(
    () => identityApi.getBackstageIdentity(),
    [identityApi],
  );

  const { value, loading, error } = useAsync(async () => {
    const apiBaseUrl = await discoveryApi.getBaseUrl('hetzner');
    const response = await fetchApi.fetch(`${apiBaseUrl}/summary`);
    if (!response.ok) {
      throw await ResponseError.fromResponse(response);
    }
    return (await response.json()) as HetznerSummaryResponse;
  }, [discoveryApi, fetchApi]);

  if (loading) {
    return <Progress />;
  }

  if (error) {
    return <ResponseErrorPanel error={error} />;
  }

  if (!value) {
    return null;
  }

  const { project, servers } = value;

  return (
    <Page themeId="tool">
      <Header title={project.title} subtitle="Hetzner Cloud overview">
        <HeaderLabel label="Owner" value={project.owner} />
        <HeaderLabel label="Lifecycle" value={project.lifecycle} />
        <HeaderLabel
          label="Signed in"
          value={identity?.userEntityRef ?? 'unknown'}
        />
      </Header>
      <Content>
        <ContentHeader title="Servers">
          <SupportButton>
            The frontend talks to the custom backend plugin described in the
            Backstage quickstart guide and displays data coming directly from
            the Hetzner Cloud API.
          </SupportButton>
        </ContentHeader>
        <Grid container spacing={3} direction="column">
          <Grid item>
            <InfoCard title="Project summary">
              <Grid container spacing={2}>
                <Grid item xs={12} md={4}>
                  <Typography variant="h2">
                    {project.totals.servers}
                  </Typography>
                  <Typography color="textSecondary">Total servers</Typography>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Typography variant="h2">
                    {project.totals.running}
                  </Typography>
                  <Typography color="textSecondary">
                    Running instances
                  </Typography>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Typography variant="h2">
                    {project.totals.datacenters.length}
                  </Typography>
                  <Typography color="textSecondary">
                    Active datacenters
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography gutterBottom>Datacenters</Typography>
                  {project.totals.datacenters.map(dc => (
                    <Chip
                      key={dc}
                      label={dc}
                      size="small"
                      style={{ marginRight: 8, marginBottom: 8 }}
                    />
                  ))}
                  {project.totals.datacenters.length === 0 && (
                    <Typography color="textSecondary">
                      No servers registered with the provider yet.
                    </Typography>
                  )}
                </Grid>
              </Grid>
            </InfoCard>
          </Grid>
          <Grid item>
            <Table
              title="Hetzner servers"
              options={{
                paging: false,
                search: true,
                padding: 'dense',
              }}
              columns={columns}
              data={servers}
            />
          </Grid>
        </Grid>
      </Content>
    </Page>
  );
};
