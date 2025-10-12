/**
 * Example 2: Multi-Section Form Layout
 * 
 * A form with multiple sections, demonstrating vertical stacking
 * with different card heights and content.
 */

import React from 'react';

interface LayoutComponents {
  Stack: any;
  Card: any;
  Space: any;
}

export const createExample2 = ({ Stack, Card, Space }: LayoutComponents) => {
  return (
    <Stack direction="vertical" gap={24} padding={20} constraints={{ minWidth: 600, maxWidth: 800 }}>
      {/* Title Section */}
      <Card
        title="New User Registration"
        content="Please fill out all required fields"
        variant="primary"
        constraints={{ preferredHeight: 100 }}
      />

      {/* Two column section */}
      <Stack direction="horizontal" gap={16}>
        <Space grow={1}>
          <Card
            title="Personal Information"
            content="Name, email, and contact details should be entered here"
            constraints={{ minHeight: 120 }}
          />
        </Space>
        <Space grow={1}>
          <Card
            title="Account Settings"
            content="Choose your username and password preferences"
            constraints={{ minHeight: 120 }}
          />
        </Space>
      </Stack>

      {/* Full width section */}
      <Card
        title="Address Information"
        content="Enter your complete mailing address including street, city, state, and postal code"
        variant="secondary"
        constraints={{ minHeight: 100 }}
      />

      {/* Three column section */}
      <Stack direction="horizontal" gap={12}>
        <Space grow={1}>
          <Card
            title="Plan"
            content="Free"
            variant="success"
          />
        </Space>
        <Space grow={1}>
          <Card
            title="Duration"
            content="Monthly"
            variant="success"
          />
        </Space>
        <Space grow={1}>
          <Card
            title="Price"
            content="$0/mo"
            variant="success"
          />
        </Space>
      </Stack>

      {/* Action buttons */}
      <Stack direction="horizontal" gap={12} distribute="end">
        <Card content="Cancel" constraints={{ preferredWidth: 120, preferredHeight: 50 }} />
        <Card
          content="Submit"
          variant="primary"
          constraints={{ preferredWidth: 120, preferredHeight: 50 }}
        />
      </Stack>
    </Stack>
  );
};

export const exampleInfo2 = {
  name: 'Multi-Section Form',
  description: 'A form with multiple sections demonstrating vertical stacking with varying card heights',
};
