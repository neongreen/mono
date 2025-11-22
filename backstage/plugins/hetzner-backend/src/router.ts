import { HttpAuthService } from '@backstage/backend-plugin-api';
import { InputError, NotFoundError } from '@backstage/errors';
import express from 'express';
import Router from 'express-promise-router';
import { hetznerServiceRef } from './services/HetznerService';

export async function createRouter({
  httpAuth,
  hetznerService,
}: {
  httpAuth: HttpAuthService;
  hetznerService: typeof hetznerServiceRef.T;
}): Promise<express.Router> {
  const router = Router();
  router.use(express.json());

  const requireCredentials = async (req: express.Request) =>
    httpAuth.credentials(req, { allow: ['user', 'service'] });

  router.get('/summary', async (req, res) => {
    await requireCredentials(req);
    res.json(await hetznerService.snapshot());
  });

  router.get('/servers', async (req, res) => {
    await requireCredentials(req);
    res.json({ servers: await hetznerService.listServers() });
  });

  router.get('/servers/:id', async (req, res) => {
    await requireCredentials(req);
    const serverId = Number(req.params.id);
    if (Number.isNaN(serverId)) {
      throw new InputError(`Invalid server id '${req.params.id}'`);
    }

    const server = await hetznerService.getServerById(serverId);
    if (!server) {
      throw new NotFoundError(`No Hetzner server found with id '${serverId}'`);
    }

    res.json(server);
  });

  return router;
}
