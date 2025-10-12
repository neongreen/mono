/**
 * Example 1: Dashboard Layout
 * 
 * A typical dashboard with a header, sidebar, and main content area
 * divided into cards.
 */

import React from 'react';

interface LayoutComponents {
  Stack: any;
  Card: any;
  Space: any;
}

export const createExample1 = ({ Stack, Card, Space }: LayoutComponents) => {
  return (
    <Stack direction="vertical" gap={16} padding={20} constraints={{ minWidth: 800, minHeight: 600 }}>
      {/* Header */}
      <Card
        title="Dashboard Header"
        content="Welcome to your dashboard"
        variant="primary"
        constraints={{ preferredHeight: 80 }}
      />

      {/* Main content area */}
      <Stack direction="horizontal" gap={16} constraints={{ minHeight: 400 }}>
        {/* Sidebar */}
        <Space grow={1} constraints={{ minWidth: 200, maxWidth: 250 }}>
          <Stack direction="vertical" gap={12}>
            <Card title="Navigation" content="Menu items go here" variant="secondary" />
            <Card title="Quick Actions" content="Common tasks" variant="secondary" />
          </Stack>
        </Space>

        {/* Content */}
        <Space grow={3}>
          <Stack direction="vertical" gap={16}>
            <Stack direction="horizontal" gap={16}>
              <Card
                title="Metric 1"
                content="Key performance indicator with some additional context"
                variant="success"
              />
              <Card
                title="Metric 2"
                content="Another important metric to track"
                variant="warning"
              />
            </Stack>
            <Card
              title="Main Content"
              content="This is the primary content area where the main information would be displayed. It should take up most of the space and be easy to read."
              constraints={{ minHeight: 200 }}
            />
          </Stack>
        </Space>
      </Stack>

      {/* Footer */}
      <Card
        content="Footer information"
        constraints={{ preferredHeight: 50 }}
      />
    </Stack>
  );
};

export const exampleInfo1 = {
  name: 'Dashboard Layout',
  description: 'A typical dashboard with header, sidebar, and main content area divided into cards',
};
