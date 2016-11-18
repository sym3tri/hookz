# hookz

Responds to webhooks by emitting kubernetes jobs.

[WIP massively incomplete]

## Build

Requirements:
- A modern version of Go
- Make
- https://github.com/Masterminds/glide

To Build the project:
``` make
```

## Dependencies

### Adding Dependencies

After adding a new `import` to the source, use `glide get` to add the dependency to the `glide.yaml` and `glide.lock` files.

```
glide get -u github.com/$ORG/$PROJ
```

### Updating Dependencies

To update an existing package, edit the `glide.yaml` file to the desired verison (most likely a git hash), and run `glide update`.

```
{{ edit the entry in glide.yaml }}
glide update -u github.com/$ORG/$PROJ
```

If the update was successful, `glide.lock` will have been updated to reflect the changes to `glide.yaml` and the package will have been updated in `vendor`.

