import { Grid, Typography } from '@material-ui/core';
import {
  EmptyState,
  InfoCard,
  StructuredMetadataTable,
} from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { HETZNER_DATA_ANNOTATION, HetznerServerSummary } from '../../types';

const formatDateTime = (isoDate: string) =>
  new Date(isoDate).toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });

export const EntityHetznerContent = () => {
  const { entity } = useEntity();
  const rawAnnotation = entity.metadata.annotations?.[HETZNER_DATA_ANNOTATION];

  if (!rawAnnotation) {
    return (
      <EmptyState
        title="This resource is not managed in Hetzner Cloud"
        missing="data"
        description="Register the resource through the Hetzner entity provider to see live metadata."
      />
    );
  }

  let parsed: HetznerServerSummary | undefined;
  try {
    parsed = JSON.parse(rawAnnotation) as HetznerServerSummary;
  } catch {
    return (
      <EmptyState
        title="Invalid Hetzner metadata"
        missing="data"
        description="Backstage could not parse the metadata stored on this entity."
      />
    );
  }

  if (!parsed) {
    return (
      <EmptyState
        title="Missing Hetzner metadata"
        missing="data"
        description="This entity does not contain valid Hetzner server information."
      />
    );
  }

  const metadata = {
    'Server type': parsed.serverType,
    Datacenter: parsed.datacenter,
    Location: parsed.location,
    'IPv4 address': parsed.ipv4Address ?? '—',
    'IPv6 address': parsed.ipv6Address ?? '—',
    Status: parsed.status,
    Locked: parsed.locked ? 'Locked' : 'Unlocked',
    'Created at': formatDateTime(parsed.createdAt),
  };

  return (
    <InfoCard title="Hetzner Cloud">
      <Grid container spacing={2} direction="column">
        <Grid item>
          <Typography variant="h6">{parsed.name}</Typography>
          <Typography variant="body2" color="textSecondary">
            Server #{parsed.id}
          </Typography>
        </Grid>
        <Grid item>
          <StructuredMetadataTable metadata={metadata} />
        </Grid>
      </Grid>
    </InfoCard>
  );
};
