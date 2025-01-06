# Minecraft Asset Uploader

This small Program Downloads Minecraft Assets from the [MCAsset Repo.](https://github.com/InventivetalentDev/minecraft-assets), filters a couple of Assets then converts them into Objects and Upload them to an S3 Object-Storage

## Usage

`./CENCAU --version=<minecraft-version>`

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

## Todo

- Configurable Filter
- Additional Export to **.zip**
- Additional Export to **.tar.gz**

# [License](/LICENSE)

Copyright 2025 RedCraft Studios

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
