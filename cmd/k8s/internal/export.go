package internal

import "epos-opensource/common"

func Export(path string) error {
	err := common.Export(path, ".env", []byte(envFile))
	if err != nil {
		return err
	}

	for name, content := range EmbeddedManifestContents {
		err = common.Export(path, name, []byte(content))
		if err != nil {
			return err
		}
		common.PrintDone("Exported %s", name)
	}

	common.PrintDone("Successfully exported default environment files in %s", path)
	return nil
}
