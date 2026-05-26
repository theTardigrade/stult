$includeExtensions = @(
	'.go',
	'.stult',
	'.json',
	'.yaml'
)

$includeFileNames = @()

$excludeDirPattern = '([\\/])(\.git|dist)([\\/]|$)'

$excludeNestedDirs = @()

$excludeNestedDirPatterns = $excludeNestedDirs | ForEach-Object {
	'([\\/])' + (($_ | ForEach-Object { [regex]::Escape($_) }) -join '[\\/]') + '([\\/]|$)'
}

$excludeFileNames = @()

Get-ChildItem -Recurse -File |
	Where-Object {
		$fullName = $_.FullName
		$extension = $_.Extension.ToLowerInvariant()
		$name = $_.Name

		$fullName -notmatch $excludeDirPattern -and
		-not ($excludeNestedDirPatterns | Where-Object { $fullName -match $_ }) -and
		$excludeFileNames -notcontains $name -and
		(
			$includeExtensions -contains $extension -or
			$includeFileNames -contains $name
		)
	} |
	Sort-Object FullName |
	ForEach-Object {
		$relativePath = Resolve-Path -Relative $_.FullName

		""
		"===== $relativePath ====="
		""
		Get-Content $_.FullName -Encoding utf8
	} |
	Set-Content all-src.txt -Encoding utf8