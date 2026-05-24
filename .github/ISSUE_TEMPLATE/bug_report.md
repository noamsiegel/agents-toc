---
name: Bug report
about: Something agents-toc did that it shouldn't have, or didn't do that it should have.
labels: bug
---

## What happened

<!-- One short paragraph. -->

## Reproducer

<!-- Minimum config + file structure + command. -->

```yaml
# .agentsmdrc.yaml
```

```
project layout
```

```
$ agents-toc <command>
<output>
```

## What you expected

## Version

```
agents-toc --version
```

## Marker-ownership invariant

agents-toc's contract is that it owns ONLY the bytes between the configured
markers. If this bug involves bytes outside the markers being touched, please
say so explicitly — it raises priority.
