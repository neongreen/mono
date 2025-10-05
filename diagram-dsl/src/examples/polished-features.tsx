import React from 'react';
import { 
  Slide, Title, Subtitle, Card, Text, Arrow,
  Cluster, Container, Group, Divider, Badge,
  List, Section, TwoColumn,
  renderToSVG 
} from '../index';
import { writeFileSync } from 'fs';

// Example 1: Enhanced Arrow Styles
const ArrowStylesSlide = () => (
  <Slide>
    <Title level={1}>Enhanced Arrow Styles</Title>
    <Subtitle>Dashed, curved, and step arrows with different head types</Subtitle>
    
    <Group label="Solid Arrows" spacing="relaxed" marginTop={20}>
      <TwoColumn gap={40}>
        <Card id="solid-1" variant="primary" width={200} height={80}>
          <Text fontSize={14} fontWeight="bold">Source</Text>
          <Text fontSize={11}>Straight arrow</Text>
        </Card>
        <Card id="solid-2" variant="secondary" width={200} height={80}>
          <Text fontSize={14} fontWeight="bold">Target</Text>
          <Text fontSize={11}>Solid line</Text>
        </Card>
      </TwoColumn>
      <Arrow from="solid-1" to="solid-2" style="solid" curve="straight" label="Solid" />
    </Group>
    
    <Group label="Dashed Arrows" spacing="relaxed" marginTop={20}>
      <TwoColumn gap={40}>
        <Card id="dashed-1" variant="success" width={200} height={80}>
          <Text fontSize={14} fontWeight="bold">Start</Text>
        </Card>
        <Card id="dashed-2" variant="warning" width={200} height={80}>
          <Text fontSize={14} fontWeight="bold">End</Text>
        </Card>
      </TwoColumn>
      <Arrow from="dashed-1" to="dashed-2" style="dashed" curve="curved" label="Dashed" color="#ff9800" />
    </Group>
    
    <Group label="Step Arrows" spacing="relaxed" marginTop={20}>
      <TwoColumn gap={40}>
        <Card id="step-1" variant="accent" width={200} height={80}>
          <Text fontSize={14} fontWeight="bold">A</Text>
        </Card>
        <Card id="step-2" variant="danger" width={200} height={80}>
          <Text fontSize={14} fontWeight="bold">B</Text>
        </Card>
      </TwoColumn>
      <Arrow from="step-1" to="step-2" style="solid" curve="step" label="Step" color="#c62828" />
    </Group>
  </Slide>
);

// Example 2: Arrow Head Types
const ArrowHeadTypesSlide = () => (
  <Slide>
    <Title level={1}>Arrow Head Types</Title>
    <Subtitle>Different arrow head and tail markers</Subtitle>
    
    <Group spacing="relaxed" marginTop={30}>
      <Group label="Arrow Heads" direction="horizontal" spacing="normal">
        <Card id="head-1" variant="primary" width={150} height={60}>
          <Text fontSize={12}>Standard</Text>
        </Card>
        <Card id="head-2" variant="primary" width={150} height={60}>
          <Text fontSize={12}>Arrow</Text>
        </Card>
        <Arrow from="head-1" to="head-2" headType="arrow" color="#1976d2" />
      </Group>
      
      <Group label="Circle Heads" direction="horizontal" spacing="normal">
        <Card id="circle-1" variant="secondary" width={150} height={60}>
          <Text fontSize={12}>Start</Text>
        </Card>
        <Card id="circle-2" variant="secondary" width={150} height={60}>
          <Text fontSize={12}>Circle</Text>
        </Card>
        <Arrow from="circle-1" to="circle-2" headType="circle" color="#7b1fa2" />
      </Group>
      
      <Group label="Diamond Heads" direction="horizontal" spacing="normal">
        <Card id="diamond-1" variant="accent" width={150} height={60}>
          <Text fontSize={12}>Begin</Text>
        </Card>
        <Card id="diamond-2" variant="accent" width={150} height={60}>
          <Text fontSize={12}>Diamond</Text>
        </Card>
        <Arrow from="diamond-1" to="diamond-2" headType="diamond" color="#f57c00" />
      </Group>
      
      <Group label="Bidirectional" direction="horizontal" spacing="normal">
        <Card id="bidir-1" variant="success" width={150} height={60}>
          <Text fontSize={12}>Two-way</Text>
        </Card>
        <Card id="bidir-2" variant="success" width={150} height={60}>
          <Text fontSize={12}>Communication</Text>
        </Card>
        <Arrow from="bidir-1" to="bidir-2" headType="arrow" tailType="arrow" color="#2e7d32" />
      </Group>
    </Group>
  </Slide>
);

// Example 3: Cluster Component
const ClusterSlide = () => (
  <Slide>
    <Title level={1}>Cluster Component</Title>
    <Subtitle>Group related elements with visual boundaries</Subtitle>
    
    <TwoColumn gap={24} marginTop={30}>
      <Cluster title="Frontend Services" variant="primary" width={500}>
        <List
          items={[
            'React Application',
            'Next.js Server',
            'Static Assets CDN'
          ]}
          fontSize={13}
        />
      </Cluster>
      
      <Cluster title="Backend Services" variant="secondary" width={500}>
        <List
          items={[
            'API Gateway',
            'Authentication Service',
            'Database Layer'
          ]}
          fontSize={13}
        />
      </Cluster>
    </TwoColumn>
    
    <Cluster 
      title="Infrastructure" 
      variant="accent" 
      width={1040} 
      marginTop={24}
    >
      <Group direction="horizontal" spacing="relaxed">
        <Badge text="AWS" variant="warning" size="medium" />
        <Badge text="Kubernetes" variant="info" size="medium" />
        <Badge text="Terraform" variant="success" size="medium" />
      </Group>
    </Cluster>
  </Slide>
);

// Example 4: Container with Dividers
const ContainerSlide = () => (
  <Slide>
    <Title level={1}>Container with Dividers</Title>
    <Subtitle>Organize content with internal sections</Subtitle>
    
    <Container
      sections={[
        {
          title: 'Configuration',
          content: React.createElement(Text, { fontSize: 12 }, 
            'Set up your project configuration and environment variables'
          )
        },
        {
          title: 'Dependencies',
          content: React.createElement(List, {
            items: ['React', 'TypeScript', 'diagram-dsl'],
            fontSize: 12
          })
        },
        {
          title: 'Scripts',
          content: React.createElement(Text, { fontSize: 12 }, 
            'Build, test, and deploy commands for your project'
          )
        }
      ]}
      variant="primary"
      width={900}
      marginTop={30}
    />
    
    <Divider width={900} marginTop={30} marginBottom={30} />
    
    <Container
      sections={[
        {
          title: 'Input',
          content: React.createElement(Text, { fontSize: 12 }, 'User data'),
          flex: 1
        },
        {
          title: 'Processing',
          content: React.createElement(Text, { fontSize: 12 }, 'Transform'),
          flex: 2
        },
        {
          title: 'Output',
          content: React.createElement(Text, { fontSize: 12 }, 'Results'),
          flex: 1
        }
      ]}
      orientation="horizontal"
      variant="secondary"
      width={900}
    />
  </Slide>
);

// Example 5: Comprehensive Example
const ComprehensiveSlide = () => (
  <Slide>
    <Title level={1}>All Features Together</Title>
    
    <Cluster title="System Architecture" variant="default" width={1080} marginTop={20}>
      <Container
        sections={[
          {
            title: 'Client Layer',
            content: React.createElement(Group, { direction: 'horizontal', spacing: 'tight' },
              React.createElement(Card, { id: 'web', variant: 'primary', width: 140, height: 50 },
                React.createElement(Text, { fontSize: 11 }, 'Web App')
              ),
              React.createElement(Card, { id: 'mobile', variant: 'primary', width: 140, height: 50 },
                React.createElement(Text, { fontSize: 11 }, 'Mobile App')
              )
            )
          },
          {
            title: 'Service Layer',
            content: React.createElement(Group, null,
              React.createElement(Card, { id: 'api', variant: 'secondary', width: 300, height: 50 },
                React.createElement(Text, { fontSize: 11 }, 'API Gateway')
              )
            )
          },
          {
            title: 'Data Layer',
            content: React.createElement(Card, { id: 'db', variant: 'accent', width: 300, height: 50 },
              React.createElement(Text, { fontSize: 11 }, 'Database')
            )
          }
        ]}
        variant="default"
        width={1000}
      />
      
      <Arrow from="web" to="api" style="solid" curve="straight" color="#1976d2" />
      <Arrow from="mobile" to="api" style="solid" curve="straight" color="#1976d2" />
      <Arrow from="api" to="db" style="dashed" curve="step" label="Query" color="#7b1fa2" />
    </Cluster>
  </Slide>
);

async function generate() {
  const examples = [
    { name: 'polished-1-arrow-styles', component: <ArrowStylesSlide /> },
    { name: 'polished-2-arrow-heads', component: <ArrowHeadTypesSlide /> },
    { name: 'polished-3-cluster', component: <ClusterSlide /> },
    { name: 'polished-4-container', component: <ContainerSlide /> },
    { name: 'polished-5-comprehensive', component: <ComprehensiveSlide /> },
  ];

  console.log('Generating polished features examples...\n');

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
