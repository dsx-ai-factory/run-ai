# async-events
`async-events` contains definitions of event structures we use as pubsub messages.

## Pubsub flow
Each struct is translated to its own stream of messages, with the same name.
When the subscriber gets the messages from the stream, he translates them back into the same struct.

## Limitations
We do don't manage a versioning mechanism.
This means, that a struct backwards compatibility cannot be broken:

- Do not remove fields from the struct
- Do not change optional fields to required or vice versa
- Do not add required fields

You can only add optional fields to the structs

## What if i need to break API?
You can create a new struct and subscribe to it.
Keep the subscription to both streams until you no longer support the old one.