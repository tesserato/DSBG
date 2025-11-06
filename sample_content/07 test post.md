---
title: "Extended Test Post: Edge Cases and Markdown Features"
description: "An extended test post to explore more edge cases and Markdown features, including citations, tables (long and short), and more. This post aims to thoroughly test formatting and rendering capabilities. Lets extend this description, for example: Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
created: 2025 02 16
coverImagePath: potato.webp
tags: [test tag 01, test tag 02, test tag 03, test tag 04, test tag 05, test tag 06, test tag 07, test tag 08, test tag 09, test tag 10, edge case, markdown feature, citation, table]
---

If you are seeing this post through the blog, you can find the source code for this post [here](https://github.com/tesserato/DSBG/blob/main/sample_content/07%20test%20post.md).

# Very long H1 heading: lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

## Very long H2 heading: lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

### Very long H3 heading: lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

#### Very long H4 heading: lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.



# Markdown Content Examples

Here's a mix of common Markdown elements to show you how they look:

## Math

Inline math (\$ ... \$): $ E = mc^2 $

Paragraph math (\\\$\\\$ ... \\\$\\\$):

\$\$ \hat{f} (\xi)=\int_{-\infty}^{\infty}f(x)e^{-2\pi ix\xi}dx \$\$

## Lists

- Unordered list item
- Another unordered list item
  - Nested unordered list item
  - Another nested unordered list item
    - Doubly nested unordered list item
    - Another doubly nested unordered list item
      - Triply nested unordered list item
      - Another triply nested unordered list item
      - Yet another triply nested unordered list item
        - Quadruply nested unordered list item
        - Another quadruply nested unordered list item
        - Yet another quadruply nested unordered list item
- Yet another list item
- Another list item

1.  Ordered list item
2.  Another ordered list item
    1.  Nested ordered list item
    2.  Another nested ordered list item
        1.  Doubly nested ordered list item
        2.  Another doubly nested ordered list item
            1.  Triply nested ordered list item
            2.  Another triply nested ordered list item
            3.  Yet another triply nested ordered list item
                1.  Quadruply nested ordered list item
                2.  Another quadruply nested ordered list item
                3.  Yet another quadruply nested ordered list item
3.  Yet another ordered list item
4.  Another ordered list item


## Code

You can include code blocks with syntax highlighting:

## Go example:
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, DSBG!")
}
```

## Rust example:
```rust
fn main() {
    println!("Hello, DSBG!");
}
```

## Python example:
```python
def main():
    print("Hello, DSBG!")

if __name__ == "__main__":
    main()
```

## JavaScript example:
```javascript
function main() {
  console.log("Hello, DSBG!");
}

main();
```

## Images

Images are automatically copied to the output directory:

![A Butterfly](01_butterfly.webp)

## Text Formatting

Here's some **bold** and *italic* text, and some `inline code`.

## Links and Quotes

Check out this [example link](https://example.com).

> This is a block quote. Useful for citing resources.

# More Markdown Features & Edge Cases

Let's push the boundaries with more features and edge cases.

## Horizontal Rules

Here are different ways to create horizontal rules:

\-\-\-

---

\*\*\*

***

\_\_\_

___

## Strikethrough

~~This text is strikethrough.~~

## Task Lists

- [x] Completed task
- [ ] Incomplete task
    - [x] Nested completed task
    - [ ] Nested incomplete task

## Emojis

Let's see if emojis render correctly: 🎉 👍 🚀 🤔

## HTML

Basic HTML should be ignored or rendered as raw HTML depending on the Markdown processor. Let's try a simple `<br>` and `<span>`.

- `This is line 1.<br>This is line 2.`

This is line 1.<br>This is line 2.

- `This is <span style="color:red;">red text</span>.`

This is <span style="color:red;">red text</span>.

## Escaping Characters

Characters that have special meaning in Markdown can be escaped with a backslash `\`.

\*asterisks\*
\_underscores\_
\#hash\#
\\backslash\\
\`backtick\`
\>greater than\>
\<less than\<
\[brackets\[
\]brackets\]
\(parentheses\(
\)parentheses\)

<!-- ## Nested Blockquotes

> This is an outer blockquote.
>
> > This is a nested blockquote.
> >
> > > And this is even more nested. -->

## Tables

### Short Table

| Header 1 | Header 2 | Header 3 |
| -------- | -------- | -------- |
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |

### Long Table

This table is designed to be wide and tall to test scrolling and wrapping behaviors.

| Very Long Header Column 1 | Very Long Header Column 2 | Very Long Header Column 3 | Very Long Header Column 4 | Very Long Header Column 5 | Very Long Header Column 6 | Very Long Header Column 7 | Very Long Header Column 8 | Very Long Header Column 9 | Very Long Header Column 10 |
| ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- | -------------------------- |
| Very Long Content Cell 1  | Very Long Content Cell 2  | Very Long Content Cell 3  | Very Long Content Cell 4  | Very Long Content Cell 5  | Very Long Content Cell 6  | Very Long Content Cell 7  | Very Long Content Cell 8  | Very Long Content Cell 9  | Very Long Content Cell 10  |
| Another Long Cell 1       | Another Long Cell 2       | Another Long Cell 3       | Another Long Cell 4       | Another Long Cell 5       | Another Long Cell 6       | Another Long Cell 7       | Another Long Cell 8       | Another Long Cell 9       | Another Long Cell 10       |
| More Content Cell 1       | More Content Cell 2       | More Content Cell 3       | More Content Cell 4       | More Content Cell 5       | More Content Cell 6       | More Content Cell 7       | More Content Cell 8       | More Content Cell 9       | More Content Cell 10       |
| And Even More Content 1   | And Even More Content 2   | And Even More Content 3   | And Even More Content 4   | And Even More Content 5   | And Even More Content 6   | And Even More Content 7   | And Even More Content 8   | And Even More Content 9   | And Even More Content 10   |
| Last Row Long Cell 1      | Last Row Long Cell 2      | Last Row Long Cell 3      | Last Row Long Cell 4      | Last Row Long Cell 5      | Last Row Long Cell 6      | Last Row Long Cell 7      | Last Row Long Cell 8      | Last Row Long Cell 9      | Last Row Long Cell 10      |
| One More Row Cell 1       | One More Row Cell 2       | One More Row Cell 3       | One More Row Cell 4       | One More Row Cell 5       | One More Row Cell 6       | One More Row Cell 7       | One More Row Cell 8       | One More Row Cell 9       | One More Row Cell 10       |


### Table with Mixed Content

| Type        | Description               | Example                               |
| ----------- | ------------------------- | ------------------------------------- |
| **Bold**    | Text with strong emphasis | `**bold text**`                       |
| *Italic*    | Text with emphasis        | `*italic text*`                       |
| `Code`      | Inline code               | `` `inline code` ``                   |
| [Link](url) | Hyperlink                 | `[example link](https://example.com)` |

## Citations and Footnotes

While Markdown doesn't have explicit citation syntax, footnotes can be used to create citations or references.[^citation-example]

Here is some text that needs a citation.[^another-citation]  And another sentence that also needs a citation.[^third-citation]

[^citation-example]: This is an example citation footnote. You can put your citation details here. For example, Author, *Title*, Publisher, Year.
[^another-citation]: Another citation footnote, perhaps to a different source.
[^third-citation]: And yet another footnote to demonstrate multiple citations.

## Very Long Words and Strings

Let's test how very long words and strings are handled, especially for layout and wrapping.

VeryLongWordWithoutSpacesToTestWordBreakingAndOverflow.

AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA

## Combinations

Let's combine some features for fun:

1.  A list item with **bold text**, *italic text*, and [a link](https://example.com).
2.  > A blockquote containing `inline code` and a [link](https://example.com) and *italic text*.
3.  A table cell with **bold text**, *italic text*, `inline code`, and [a link](https://example.com).

| Combination Type | Example Content                                                                                   |
| ---------------- | ------------------------------------------------------------------------------------------------- |
| List Item        | 1. A list item with **bold text**, *italic text*, and [a link](https://example.com).              |
| Table Cell       | A table cell with **bold text**, *italic text*, `inline code`, and [a link](https://example.com). |

This extended test post should cover a wider range of Markdown features and edge cases, providing a more thorough test of rendering and formatting.