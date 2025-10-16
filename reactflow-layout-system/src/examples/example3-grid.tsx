/**
 * Example 3: Complex Grid Layout
 * 
 * A more complex layout with nested stacks showing how the system
 * handles deep nesting and varied content.
 */

import React from 'react';

interface LayoutComponents {
  Stack: any;
  Card: any;
  Space: any;
}

export const createExample3 = ({ Stack, Card, Space }: LayoutComponents) => {
  return (
    <Stack direction="vertical" gap={20} padding={20} constraints={{ minWidth: 900 }}>
      {/* Top row with title */}
      <Card
        title="Project Overview"
        content="Comprehensive view of all project components and their relationships"
        variant="primary"
        constraints={{ preferredHeight: 90 }}
      />

      {/* Main content grid */}
      <Stack direction="horizontal" gap={16}>
        {/* Left column - tall cards */}
        <Space grow={2}>
          <Stack direction="vertical" gap={16}>
            <Card
              title="Team Members"
              content="Active contributors and their roles in the project"
              variant="secondary"
              constraints={{ minHeight: 150 }}
            />
            <Card
              title="Recent Activity"
              content="Latest updates, commits, and changes made to the project"
              variant="secondary"
              constraints={{ minHeight: 150 }}
            />
          </Stack>
        </Space>

        {/* Middle column - grid of smaller cards */}
        <Space grow={3}>
          <Stack direction="vertical" gap={12}>
            <Stack direction="horizontal" gap={12}>
              <Card title="Issues" content="23 open" variant="warning" />
              <Card title="PRs" content="5 pending" variant="primary" />
              <Card title="Builds" content="Passing" variant="success" />
            </Stack>
            <Card
              title="Statistics"
              content="Detailed analytics and metrics for project health monitoring"
              constraints={{ minHeight: 120 }}
            />
            <Stack direction="horizontal" gap={12}>
              <Card title="Coverage" content="87%" variant="success" />
              <Card title="Tests" content="432 passing" variant="success" />
            </Stack>
          </Stack>
        </Space>

        {/* Right column - info cards */}
        <Space grow={2}>
          <Stack direction="vertical" gap={16}>
            <Card
              title="Documentation"
              content="Guides and API references"
              constraints={{ minHeight: 100 }}
            />
            <Card
              title="Resources"
              content="Links to external tools and services"
              constraints={{ minHeight: 100 }}
            />
            <Card
              title="Support"
              content="Get help from the community"
              constraints={{ minHeight: 100 }}
            />
          </Stack>
        </Space>
      </Stack>

      {/* Bottom status bar */}
      <Stack direction="horizontal" gap={12}>
        <Card content="Last updated: 2 minutes ago" constraints={{ preferredHeight: 60 }} />
        <Card content="Status: Active" variant="success" constraints={{ preferredHeight: 60 }} />
      </Stack>
    </Stack>
  );
};

export const exampleInfo3 = {
  name: 'Complex Grid Layout',
  description: 'A complex layout with nested stacks showing deep nesting and varied content',
};
