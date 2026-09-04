# CardItem Component

A generic, reusable card item component that provides a consistent card-based layout with optional delete functionality.

## Features

- **Flexible Delete Button**: Optional delete button with customizable icon, tooltip, and permission-based visibility
- **Slot-based Content**: Uses Vue slots for maximum flexibility (`summary`, `content`, `footer`)
- **Customizable Styling**: Configurable heights, margins, and collapsed states
- **Permission Integration**: Built-in support for permission-based button visibility

## Usage Examples

### Basic Usage

```vue
<template>
  <card-item @delete="handleDelete">
    <template #summary>
      <div>Item Summary</div>
    </template>
  </card-item>
</template>
```

### With Permission-based Delete Button

```vue
<template>
  <card-item
    :permission-config="{ resourceType: ResourceType.Projects, action: Action.Delete }"
    :delete-button-tooltip="'Delete project'"
    @delete="handleDelete"
  >
    <template #summary>
      <div>Project Summary</div>
    </template>
  </card-item>
</template>
```

### With Content and Footer

```vue
<template>
  <card-item :collapsed="false" content-height="400px" @delete="handleDelete">
    <template #summary>
      <div>Item Summary</div>
    </template>
    <template #content>
      <div>Detailed content here</div>
    </template>
    <template #footer>
      <q-btn label="Save" @click="save" />
      <q-btn label="Cancel" @click="cancel" />
    </template>
  </card-item>
</template>
```

### Without Delete Button

```vue
<template>
  <card-item :show-delete-button="false">
    <template #summary>
      <div>Read-only item</div>
    </template>
  </card-item>
</template>
```

### With Tight Spacing (for lists)

```vue
<template>
  <card-item :tight-spacing="true" @delete="handleDelete">
    <template #summary>
      <div>List item with no gaps</div>
    </template>
  </card-item>
</template>
```

### With Regular Spacing (for cards)

```vue
<template>
  <card-item @delete="handleDelete">
    <template #summary>
      <div>Card item with spacing</div>
    </template>
  </card-item>
</template>
```

## Props

| Prop                   | Type    | Default                   | Description                                           |
| ---------------------- | ------- | ------------------------- | ----------------------------------------------------- |
| `showDeleteButton`     | Boolean | `true`                    | Whether to show the delete button                     |
| `disabledDeleteButton` | Boolean | `false`                   | Whether the delete button is disabled                 |
| `deleteIcon`           | String  | `"mdi-trash-can-outline"` | Icon for the delete button                            |
| `deleteButtonTooltip`  | String  | `""`                      | Tooltip text for the delete button                    |
| `permissionConfig`     | Object  | `null`                    | Permission configuration for delete button visibility |
| `minHeight`            | String  | `"50px"`                  | Minimum height of the card                            |
| `contentHeight`        | String  | `"280px"`                 | Height of content when expanded                       |
| `collapsed`            | Boolean | `false`                   | Whether the item is in collapsed state                |
| `marginTop`            | Boolean | `false`                   | Whether to add top margin                             |
| `tightSpacing`         | Boolean | `false`                   | Whether to use tight spacing (no gaps between items)  |

## Events

| Event    | Description                               |
| -------- | ----------------------------------------- |
| `delete` | Emitted when the delete button is clicked |

## Slots

| Slot      | Description                            |
| --------- | -------------------------------------- |
| `summary` | Main content displayed in the card     |
| `content` | Additional content (optional)          |
| `footer`  | Footer content with buttons (optional) |
