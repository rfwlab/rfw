# Events

rfw offers two layers of events:

1. **DOM events** – listen to browser events via helpers like
   `events.Listen("click", el)` which return Go channels.
2. **Store events** – reactive stores trigger updates when their values
   change, allowing components to respond without manual listeners.

DOM events are ideal for user interactions while store events model
application state. Both approaches keep the amount of handwritten
JavaScript to a minimum.
The example reacts to user interactions.

@include:ExampleFrame:{code:"/examples/components/event_component.go", uri:"/examples/event"}
