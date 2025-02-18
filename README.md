# Minecraft Asset Uploader

This small Program Downloads Minecraft Assets from the [MCAsset Repo.](https://github.com/InventivetalentDev/minecraft-assets), filters a couple of Assets then converts them into Objects and Upload them to an S3 Object-Storage

### Table of Contents

- [CLI](#cli)
- [Run](#run)
- [Example .env file](#example-env-file)
- [TODO](#todo)
- [License](#license)
- [Contributers](#contributers)

## CLI

```
assetuploader
    --version
        Set's the minecraft version
```

## Run

```bash
go run cmd/assetuploader/main.go <arguments>
```

## Example .env file

```
S3_BUCKET=<bucket-name>
S3_PREFIX=<key-prefix>
S3_ENDPOINT=<endpoint>
S3_REGION=<region>
S3_ACCESS_KEY=<access-key>
S3_SECRET_KEY=<secret-key>
S3_PATH_STYLE=<true | false>
```

## TODO

- Configurable Filter
- Additional Export to **.zip**
- Additional Export to **.tar.gz**

## License

[PolarLite](https://github.com/CraftUniverse/PolarLite) © 2025 by [CraftUniverse](https://github.com/CraftUniverse) is licensed under [Creative Commons Attribution-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-sa/4.0/?ref=chooser-v1)

## Contributers

- [@Turboman3000](https://github.com/Turboman3000)
