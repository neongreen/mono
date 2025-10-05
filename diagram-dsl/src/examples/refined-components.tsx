import React from 'react';
import { 
  Slide, Title, Subtitle, Text,
  Panel, Well, Icon, Steps, Cluster, Container,
  List, Card, Badge, Divider,
  renderToSVG 
} from '../index';
import { writeFileSync } from 'fs';

// Example 1: Panel Component
const PanelExample = () => (
  <Slide>
    <Title level={1}>Panel Component</Title>
    <Subtitle>Structured containers with header and footer</Subtitle>
    
    <Panel
      header={
        <Text fontSize={16} fontWeight="bold">User Profile</Text>
      }
      footer={
        <Text fontSize={12} color="#666">Last updated: Today</Text>
      }
      variant="primary"
      width={900}
      marginTop={30}
    >
      <List
        items={[
          'Name: John Doe',
          'Email: john@example.com',
          'Role: Administrator'
        ]}
        fontSize={14}
      />
    </Panel>
    
    <Panel
      header={
        <Text fontSize={16} fontWeight="bold">Settings</Text>
      }
      variant="secondary"
      width={900}
      marginTop={24}
      elevation={2}
    >
      <Text fontSize={14}>Configure your application settings here</Text>
    </Panel>
  </Slide>
);

// Example 2: Well Component
const WellExample = () => (
  <Slide>
    <Title level={1}>Well Component</Title>
    <Subtitle>Inset containers for secondary content</Subtitle>
    
    <Text fontSize={16} marginTop={30} marginBottom={20}>
      Main content goes here
    </Text>
    
    <Well variant="info" width={900}>
      <Text fontSize={14} fontWeight="bold" marginBottom={8}>ℹ️ Information</Text>
      <Text fontSize={13}>
        This is additional context provided in an info well.
      </Text>
    </Well>
    
    <Spacer size={20} />
    
    <Well variant="success" width={900}>
      <Text fontSize={14} fontWeight="bold" marginBottom={8}>✓ Success</Text>
      <Text fontSize={13}>
        Operation completed successfully!
      </Text>
    </Well>
    
    <Spacer size={20} />
    
    <Well variant="warning" width={900}>
      <Text fontSize={14} fontWeight="bold" marginBottom={8}>⚠️ Warning</Text>
      <Text fontSize={13}>
        Please review before proceeding.
      </Text>
    </Well>
  </Slide>
);

// Example 3: Icon Component
const IconExample = () => (
  <Slide>
    <Title level={1}>Icon Component</Title>
    <Subtitle>Visual symbols and indicators</Subtitle>
    
    <Group direction="horizontal" spacing="relaxed" marginTop={40}>
      <Stack gap={12} alignItems="center">
        <Icon symbol="✓" size="large" color="#2e7d32" circular backgroundColor="#e8f5e9" />
        <Text fontSize={12}>Success</Text>
      </Stack>
      
      <Stack gap={12} alignItems="center">
        <Icon symbol="✗" size="large" color="#c62828" circular backgroundColor="#ffebee" />
        <Text fontSize={12}>Error</Text>
      </Stack>
      
      <Stack gap={12} alignItems="center">
        <Icon symbol="⚠" size="large" color="#ff9800" circular backgroundColor="#fff3e0" />
        <Text fontSize={12}>Warning</Text>
      </Stack>
      
      <Stack gap={12} alignItems="center">
        <Icon symbol="ℹ" size="large" color="#1976d2" circular backgroundColor="#e3f2fd" />
        <Text fontSize={12}>Info</Text>
      </Stack>
      
      <Stack gap={12} alignItems="center">
        <Icon symbol="★" size="large" color="#f57c00" circular backgroundColor="#fff3e0" />
        <Text fontSize={12}>Favorite</Text>
      </Stack>
    </Group>
    
    <Divider width={900} marginTop={40} marginBottom={40} />
    
    <Group direction="horizontal" spacing="relaxed">
      <Icon symbol="🚀" size="xlarge" />
      <Icon symbol="💡" size="large" />
      <Icon symbol="📊" size="medium" />
      <Icon symbol="⚙️" size="small" />
    </Group>
  </Slide>
);

// Example 4: Steps Component
const StepsExample = () => (
  <Slide>
    <Title level={1}>Steps Component</Title>
    <Subtitle>Process visualization with status indicators</Subtitle>
    
    <Steps
      steps={[
        {
          number: 1,
          title: 'Installation',
          description: 'Install diagram-dsl and dependencies',
          status: 'complete'
        },
        {
          number: 2,
          title: 'Configuration',
          description: 'Set up your presentation structure',
          status: 'complete'
        },
        {
          number: 3,
          title: 'Development',
          description: 'Create your slides using components',
          status: 'active'
        },
        {
          number: 4,
          title: 'Generation',
          description: 'Generate SVG outputs',
          status: 'pending'
        },
        {
          number: 5,
          title: 'Deployment',
          description: 'Deploy your presentation',
          status: 'pending'
        }
      ]}
      width={700}
      marginTop={40}
    />
  </Slide>
);

// Example 5: Comprehensive Polish Example
const ComprehensiveExample = () => (
  <Slide>
    <Title level={1}>Polished Components</Title>
    
    <Cluster title="Architecture Overview" variant="primary" width={1080} marginTop={20}>
      <Container
        sections={[
          {
            title: 'Frontend',
            content: React.createElement(Group, { direction: 'horizontal', spacing: 'tight' },
              React.createElement(Badge, { text: 'React', variant: 'primary', size: 'small' }),
              React.createElement(Badge, { text: 'TypeScript', variant: 'info', size: 'small' })
            )
          },
          {
            title: 'Backend',
            content: React.createElement(Group, { direction: 'horizontal', spacing: 'tight' },
              React.createElement(Badge, { text: 'Node.js', variant: 'success', size: 'small' }),
              React.createElement(Badge, { text: 'GraphQL', variant: 'secondary', size: 'small' })
            )
          },
          {
            title: 'Infrastructure',
            content: React.createElement(Group, { direction: 'horizontal', spacing: 'tight' },
              React.createElement(Badge, { text: 'AWS', variant: 'warning', size: 'small' }),
              React.createElement(Badge, { text: 'Docker', variant: 'info', size: 'small' })
            )
          }
        ]}
        variant="default"
        width={1020}
      />
    </Cluster>
    
    <Divider width={1080} marginTop={24} marginBottom={24} />
    
    <Well variant="info" width={1080}>
      <Text fontSize={14} fontWeight="bold" marginBottom={8}>
        💡 These components work together seamlessly
      </Text>
      <Text fontSize={13}>
        Combine Panel, Well, Icon, Steps, Cluster, and Container to create
        professional, polished presentations with minimal effort.
      </Text>
    </Well>
  </Slide>
);

// Export missing components from imports
const { Group, Spacer, Stack } = await import('../index');

async function generate() {
  const examples = [
    { name: 'refined-1-panel', component: <PanelExample /> },
    { name: 'refined-2-well', component: <WellExample /> },
    { name: 'refined-3-icon', component: <IconExample /> },
    { name: 'refined-4-steps', component: <StepsExample /> },
    { name: 'refined-5-comprehensive', component: <ComprehensiveExample /> },
  ];

  console.log('Generating refined components examples...\n');

  for (const example of examples) {
    const svg = await renderToSVG(example.component, {
      width: 1200,
      height: 800,
      backgroundColor: 'white',
    });
    
    const filename = `${example.name}.svg`;
    writeFileSync(filename, svg);
    console.log(`✓ Generated ${filename}`);
  }
  
  console.log('\nDone!');
}

generate();
