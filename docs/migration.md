# Migration into `pluginctl`

Migrating into `pluginctl` is easy! Follow these steps to get started:

1. **Install `pluginctl`**: Make sure you have `pluginctl` installed. You can install it from source, use the pre-built binaries or `go install`. Follow the [README](../README.md) for installation instructions.

2. **Update Your Plugin Structure**: Ensure you follow a proper directory structure for your plugin. Refer to the [Mattermost plugin structure](https://developers.mattermost.com/integrate/plugins/) documentation for guidance. You probably based your plugin on the [starter template](https://github.com/mattermost/mattermost-plugin-starter-template) so this should already be in place.

3. **Migrate your custom development logic**: If you have customization to the base `Makefile`, is best you make a backup or you put the targets into a `build/custom_xxx.mk`, since the main `Makefile` is going to be overwritten by `pluginctl` in the step below. You are going to be able to see any changes in your version control system afterwards.

4. **Update the assets with `pluginctl`**: Run `pluginctl updateassets`. This will replace and create new files in your plugin.

5. **Manually remove old asset files**: You can remove the following files:
   - `build/manifest/*`: Tool gets replaced by this one (`pluginctl manifest`).
   - `build/pluginctl/*`: Tool gets replaced by this one.
   - `build/bin/*`: Not needed anymore since the local tools are being removed.
   - `build/setup.mk`: This has been moved to `build/_setup.mk`.

6. **Done!** Everything should be working now. Check your plugin's functionality to ensure everything is in order or try building/deploying it with `pluginctl` directly. Use `pluginctl help` for more information.

If you encounter any issues during the migration process, feel free to reach out for help!
