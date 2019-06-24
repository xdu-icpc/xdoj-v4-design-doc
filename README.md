# XDOJ v4 Design

## What happened to XDOJ v3?

It failed because we reinvented too many wheels and these wheels don't
work well.

## Systemd Unit Hierarchy

Now we will use [systemd](https://github.com/systemd/systemd), the System
and Service Manager to simplify our work.  A simple proof of concept is in
the `prototype` directory.  It can be tested on a system booted with systemd
with:

```
go build
sudo ./prototype
```

We plan to use the following hierarchy for the final approach:

```
-.slice
  xdojv4.slice
    xdojv4-runner0.slice
      xdojv4-solution-${uuid}.service
      xdojv4-checker-${uuid}.service
    xdojv4-runner1.slice
    ... ...
```
