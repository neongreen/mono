# Before and After: Semantic Components

This document shows real code comparisons demonstrating how semantic components reduce code and improve maintainability.

## Example 1: Simple Flowchart Box

### Before (Manual Styling)

```tsx
<Box
  id="start"
  width={200}
  height={80}
  backgroundColor="#e0f7fa"
  borderColor="#00acc1"
  borderWidth={2}
  borderRadius={8}
  padding={20}
  justifyContent="center"
  alignItems="center"
>
  <Text fontSize={18}>Start Process</Text>
</Box>
```

**Lines of code:** 14  
**Manual properties:** 9 (colors, sizes, spacing, alignment)

### After (Semantic Components)

```tsx
<Card
  id="start"
  variant="info"
  width={200}
  height={80}
>
  <Label size="lg">Start Process</Label>
</Card>
```

**Lines of code:** 8  
**Manual properties:** 2 (just width/height)  
**Code reduction:** 43%

### What Changed

- ❌ Removed manual `backgroundColor`, `borderColor`, `borderWidth`, `borderRadius`
- ❌ Removed manual `padding`, `justifyContent`, `alignItems`
- ❌ Removed manual `fontSize` specification
- ✅ Added semantic `variant="info"` (auto colors)
- ✅ Added semantic `Label size="lg"` (auto font size)

---

## Example 2: Box with Rich Text

### Before (Manual Styling)

```tsx
<Box
  id="api"
  width={180}
  height={100}
  backgroundColor="#e3f2fd"
  borderColor="#1976d2"
  borderWidth={2}
  borderRadius={8}
  padding={15}
  justifyContent="center"
  alignItems="center"
>
  <Text fontSize={16} fontWeight="bold">API Gateway</Text>
  <Text fontSize={12} color="#757575">REST/GraphQL</Text>
</Box>
```

**Lines of code:** 16  
**Manual properties:** 11 (colors, sizes, weights)  
**Issues:** Text spacing not explicit, color hardcoded

### After (Semantic Components)

```tsx
<Card
  id="api"
  variant="primary"
  width={180}
  height={100}
>
  <Stack gap={6} alignItems="center">
    <Label bold size="lg">API Gateway</Label>
    <Subtitle>REST/GraphQL</Subtitle>
  </Stack>
</Card>
```

**Lines of code:** 12  
**Manual properties:** 2 (just width/height)  
**Code reduction:** 25%

### What Changed

- ❌ Removed 5 manual styling properties from Box
- ❌ Removed manual `fontSize`, `fontWeight`, `color` from Text
- ✅ Added semantic `variant="primary"` (auto colors)
- ✅ Added explicit `Stack gap={6}` (clear spacing)
- ✅ Used `Label bold size="lg"` (semantic styling)
- ✅ Used `Subtitle` (auto gray color, smaller size)

---

## Example 3: Complete Diagram Header

### Before (Manual Styling)

```tsx
<Stack gap={20} padding={40} alignItems="center">
  <Text fontSize={32} fontWeight="bold" textAlign="center">
    Three-Tier Architecture
  </Text>
  <Text fontSize={14} color="#757575" textAlign="center">
    Modern web application design
  </Text>
  {/* ... rest of diagram */}
</Stack>
```

**Lines of code:** 10  
**Manual properties:** 7

### After (Semantic Components)

```tsx
<Stack gap={32} padding={40}>
  <Stack gap={8} alignItems="center">
    <Title level={1}>Three-Tier Architecture</Title>
    <Subtitle size="base">Modern web application design</Subtitle>
  </Stack>
  {/* ... rest of diagram */}
</Stack>
```

**Lines of code:** 7  
**Manual properties:** 0 (all semantic)  
**Code reduction:** 30%

### What Changed

- ❌ Removed manual `fontSize`, `fontWeight`, `textAlign` from title
- ❌ Removed manual `fontSize`, `color`, `textAlign` from subtitle
- ✅ Added semantic `Title level={1}` (auto 36px, bold, centered)
- ✅ Added semantic `Subtitle size="base"` (auto 14px, gray, centered)
- ✅ Added explicit `Stack gap={8}` for clear hierarchy

---

## Example 4: Full Flowchart

### Before (Manual Styling) - 59 lines

```tsx
const OldFlowchart = () => (
  <Stack gap={20} padding={40} alignItems="center" width={800} height={600}>
    <Text fontSize={32} fontWeight="bold">Simple Flowchart</Text>
    
    <Box
      id="start"
      width={200}
      height={80}
      backgroundColor="#e0f7fa"
      borderColor="#00acc1"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>Start Process</Text>
    </Box>

    <Box
      id="process"
      width={200}
      height={80}
      backgroundColor="#fff3e0"
      borderColor="#fb8c00"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>Process Data</Text>
    </Box>

    <Box
      id="end"
      width={200}
      height={80}
      backgroundColor="#f1f8e9"
      borderColor="#7cb342"
      borderWidth={2}
      borderRadius={8}
      padding={20}
      justifyContent="center"
      alignItems="center"
    >
      <Text fontSize={18}>End Process</Text>
    </Box>

    <Arrow from="start" to="process" color="#00acc1" strokeWidth={2} />
    <Arrow from="process" to="end" color="#fb8c00" strokeWidth={2} />
  </Stack>
);
```

### After (Semantic Components) - 34 lines

```tsx
const NewFlowchart = () => (
  <Stack gap={24} padding={40} alignItems="center">
    <Title level={1}>Simple Flowchart</Title>
    
    <Card id="start" variant="info" width={200} height={80}>
      <Label size="lg">Start Process</Label>
    </Card>

    <Card id="process" variant="warning" width={200} height={80}>
      <Label size="lg">Process Data</Label>
    </Card>

    <Card id="end" variant="success" width={200} height={80}>
      <Label size="lg">End Process</Label>
    </Card>

    <Arrow from="start" to="process" color="#00acc1" strokeWidth={2} />
    <Arrow from="process" to="end" color="#7cb342" strokeWidth={2} />
  </Stack>
);
```

**Code reduction:** 42% (25 lines saved)  
**Readability:** Much clearer semantic intent

---

## Key Benefits Summary

### 1. Less Code
- **Average reduction:** 30-45% fewer lines
- **Fewer properties:** 70-90% fewer manual styling props
- **Faster to write:** Semantic components are quicker to type

### 2. Better Maintainability
- **No hardcoded colors:** Use variants instead
- **Consistent styling:** Theme ensures uniformity
- **Clear intent:** `variant="success"` vs `backgroundColor="#f1f8e9"`

### 3. Rich Text Support
**Before:** Difficult to add subtitle to a box
```tsx
<Box padding={15}>
  <Text fontSize={16} fontWeight="bold">Title</Text>
  <Text fontSize={12} color="#757575">Subtitle</Text>  {/* Spacing? */}
</Box>
```

**After:** Easy and explicit
```tsx
<Card>
  <Stack gap={6}>
    <Label bold size="lg">Title</Label>
    <Subtitle>Subtitle</Subtitle>
  </Stack>
</Card>
```

### 4. Professional Defaults
- Colors from curated palette
- Typography scale (not random sizes)
- Consistent spacing (4px grid)
- Proper border radius
- Good padding by default

---

## Real-World Impact

**Diagram complexity remains the same**, but:
- ✅ 40% less typing
- ✅ No color/size decisions needed
- ✅ Professional look guaranteed
- ✅ Easy to create rich text elements
- ✅ Consistent across all diagrams
- ✅ Easier to maintain and update
