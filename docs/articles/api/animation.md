# animation

Helpers for driving simple animations and for leveraging the Web Animations API.

- `Translate(sel, from, to, dur)` moves elements.
- `Fade(sel, from, to, dur)` adjusts opacity.
- `Scale(sel, from, to, dur)` scales elements.
- `ColorCycle(sel, colors, dur)` cycles background colors.
- `Keyframes(sel, frames, opts)` runs a Web Animations API sequence and
  returns the underlying `Animation` object.

## Usage

Call the animation helpers with a CSS selector, starting and ending values,
and a duration. They can be invoked from event handlers registered in the DOM.
`Keyframes` accepts arrays of frame definitions and option maps that are
passed directly to the browser's `Element.animate` API.

The following example animates elements using these helpers.

@include:ExampleFrame:{code:"/examples/components/animation_component.go", uri:"/examples/animations"}
