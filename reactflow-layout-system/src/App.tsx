import React from 'react';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { ExamplePage } from './showcase/ExamplePage';
import { createExample1, exampleInfo1 } from './examples/example1-dashboard';
import { createExample2, exampleInfo2 } from './examples/example2-form';
import { createExample3, exampleInfo3 } from './examples/example3-grid';
import './App.css';

// Source code strings for display
const sourceCode1 = `import React from 'react';

// Dashboard Layout
<Stack direction="vertical" gap={16} padding={20}>
  <Card
    title="Dashboard Header"
    content="Welcome to your dashboard"
    variant="primary"
    constraints={{ preferredHeight: 80 }}
  />

  <Stack direction="horizontal" gap={16}>
    <Space grow={1} constraints={{ minWidth: 200, maxWidth: 250 }}>
      <Stack direction="vertical" gap={12}>
        <Card title="Navigation" content="Menu items" variant="secondary" />
        <Card title="Quick Actions" content="Common tasks" variant="secondary" />
      </Stack>
    </Space>

    <Space grow={3}>
      <Stack direction="vertical" gap={16}>
        <Stack direction="horizontal" gap={16}>
          <Card title="Metric 1" content="KPI" variant="success" />
          <Card title="Metric 2" content="Another metric" variant="warning" />
        </Stack>
        <Card title="Main Content" content="Primary content area" />
      </Stack>
    </Space>
  </Stack>

  <Card content="Footer information" constraints={{ preferredHeight: 50 }} />
</Stack>`;

const sourceCode2 = `import React from 'react';

// Multi-Section Form Layout
<Stack direction="vertical" gap={24} padding={20}>
  <Card
    title="New User Registration"
    content="Please fill out all required fields"
    variant="primary"
  />

  <Stack direction="horizontal" gap={16}>
    <Space grow={1}>
      <Card title="Personal Information" content="Name, email, contact" />
    </Space>
    <Space grow={1}>
      <Card title="Account Settings" content="Username and password" />
    </Space>
  </Stack>

  <Card
    title="Address Information"
    content="Complete mailing address"
    variant="secondary"
  />

  <Stack direction="horizontal" gap={12}>
    <Space grow={1}><Card title="Plan" content="Free" variant="success" /></Space>
    <Space grow={1}><Card title="Duration" content="Monthly" variant="success" /></Space>
    <Space grow={1}><Card title="Price" content="$0/mo" variant="success" /></Space>
  </Stack>

  <Stack direction="horizontal" gap={12} distribute="end">
    <Card content="Cancel" />
    <Card content="Submit" variant="primary" />
  </Stack>
</Stack>`;

const sourceCode3 = `import React from 'react';

// Complex Grid Layout
<Stack direction="vertical" gap={20} padding={20}>
  <Card title="Project Overview" content="Comprehensive view" variant="primary" />

  <Stack direction="horizontal" gap={16}>
    <Space grow={2}>
      <Stack direction="vertical" gap={16}>
        <Card title="Team Members" content="Active contributors" variant="secondary" />
        <Card title="Recent Activity" content="Latest updates" variant="secondary" />
      </Stack>
    </Space>

    <Space grow={3}>
      <Stack direction="vertical" gap={12}>
        <Stack direction="horizontal" gap={12}>
          <Card title="Issues" content="23 open" variant="warning" />
          <Card title="PRs" content="5 pending" variant="primary" />
          <Card title="Builds" content="Passing" variant="success" />
        </Stack>
        <Card title="Statistics" content="Detailed analytics" />
        <Stack direction="horizontal" gap={12}>
          <Card title="Coverage" content="87%" variant="success" />
          <Card title="Tests" content="432 passing" variant="success" />
        </Stack>
      </Stack>
    </Space>

    <Space grow={2}>
      <Stack direction="vertical" gap={16}>
        <Card title="Documentation" content="Guides and API" />
        <Card title="Resources" content="External tools" />
        <Card title="Support" content="Community help" />
      </Stack>
    </Space>
  </Stack>

  <Stack direction="horizontal" gap={12}>
    <Card content="Last updated: 2 min ago" />
    <Card content="Status: Active" variant="success" />
  </Stack>
</Stack>`;

const HomePage: React.FC = () => {
  return (
    <div className="home-page">
      <div className="hero">
        <h1>ReactFlow Layout System Showcase</h1>
        <p className="subtitle">
          A constraint-based layout system for building beautiful, responsive documents
        </p>
        <p className="description">
          This showcase demonstrates three different approaches to implementing a layout system
          with the same high-level API. Each approach has different trade-offs in terms of
          flexibility, complexity, and behavior.
        </p>
      </div>

      <div className="features">
        <div className="feature">
          <h3>🎯 Constraint-Based</h3>
          <p>
            Specify minimum, maximum, and preferred sizes. The system handles the rest,
            ensuring text never overflows and layouts look professional.
          </p>
        </div>
        <div className="feature">
          <h3>📦 Simple API</h3>
          <p>
            Use Stack, Card, and Space components to build complex layouts without
            manual calculations or pixel-perfect positioning.
          </p>
        </div>
        <div className="feature">
          <h3>🔄 Multiple Approaches</h3>
          <p>
            Compare Flexbox-based, Grid-based, and Constrained Absolute implementations
            to see which works best for your use case.
          </p>
        </div>
      </div>

      <div className="examples-list">
        <h2>Examples</h2>
        <div className="example-cards">
          <Link to="/example1" className="example-card">
            <h3>{exampleInfo1.name}</h3>
            <p>{exampleInfo1.description}</p>
          </Link>
          <Link to="/example2" className="example-card">
            <h3>{exampleInfo2.name}</h3>
            <p>{exampleInfo2.description}</p>
          </Link>
          <Link to="/example3" className="example-card">
            <h3>{exampleInfo3.name}</h3>
            <p>{exampleInfo3.description}</p>
          </Link>
        </div>
      </div>
    </div>
  );
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div className="app">
        <nav className="navbar">
          <Link to="/" className="logo">
            ReactFlow Layout System
          </Link>
          <div className="nav-links">
            <Link to="/example1">Dashboard</Link>
            <Link to="/example2">Form</Link>
            <Link to="/example3">Grid</Link>
          </div>
        </nav>

        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route
            path="/example1"
            element={
              <ExamplePage
                createExample={createExample1}
                exampleInfo={exampleInfo1}
                sourceCode={sourceCode1}
              />
            }
          />
          <Route
            path="/example2"
            element={
              <ExamplePage
                createExample={createExample2}
                exampleInfo={exampleInfo2}
                sourceCode={sourceCode2}
              />
            }
          />
          <Route
            path="/example3"
            element={
              <ExamplePage
                createExample={createExample3}
                exampleInfo={exampleInfo3}
                sourceCode={sourceCode3}
              />
            }
          />
        </Routes>
      </div>
    </BrowserRouter>
  );
};

export default App;
