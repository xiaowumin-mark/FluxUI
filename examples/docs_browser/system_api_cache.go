package main

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	docsSystemIconPathOnce sync.Once
	docsSystemIconPath     string

	docsSystemDefaultDialogDirOnce  sync.Once
	docsSystemDefaultDialogDirValue string

	docsDragSampleFileOnce sync.Once
	docsDragSampleFilePath string

	docsSystemRegistrationTargetsOnce  sync.Once
	docsSystemRegistrationTargetsValue docsSystemRegistrationTargets
	docsSystemRegistrationTargetsErr   error
)

func cachedDocsSystemNotificationIconPath() string {
	docsSystemIconPathOnce.Do(func() {
		candidates := []string{
			filepath.Join("examples", "docs_browser", "docs.ico"),
			filepath.Join("examples", "system_showcase", "system_showcase.ico"),
			filepath.Join("examples", "assets", "sample.ico"),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				docsSystemIconPath = candidate
				return
			}
		}
	})
	return docsSystemIconPath
}

func cachedDocsSystemDefaultDialogDir() string {
	docsSystemDefaultDialogDirOnce.Do(func() {
		if docsRoot, err := resolveDocsRootDir(); err == nil {
			docsSystemDefaultDialogDirValue = docsRoot
		}
	})
	return docsSystemDefaultDialogDirValue
}

func cachedDocsDragSampleFile() string {
	docsDragSampleFileOnce.Do(func() {
		if docsRoot, err := resolveDocsRootDir(); err == nil {
			if path, err := filepath.Abs(filepath.Join(docsRoot, "README.md")); err == nil {
				docsDragSampleFilePath = path
				return
			}
			docsDragSampleFilePath = filepath.Join(docsRoot, "README.md")
			return
		}
		path, err := filepath.Abs(filepath.Join("docs", "README.md"))
		if err != nil {
			docsDragSampleFilePath = filepath.Join("docs", "README.md")
			return
		}
		docsDragSampleFilePath = path
	})
	return docsDragSampleFilePath
}

func cachedDocsSystemRegistrationTargetsForCurrentExe() (docsSystemRegistrationTargets, error) {
	docsSystemRegistrationTargetsOnce.Do(func() {
		docsSystemRegistrationTargetsValue, docsSystemRegistrationTargetsErr = buildDocsSystemRegistrationTargetsForCurrentExe()
	})
	return docsSystemRegistrationTargetsValue, docsSystemRegistrationTargetsErr
}
