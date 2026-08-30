#!/bin/bash

tagsPublished=$(yq e '.x-tagGroups[].tags'  $REDOC_ROOT_API_PATH | sed 's/- //g' | sort )

# Get parent group names from x-tagGroups (these are category headers, not endpoint tags)
parentGroupTags=$(yq e '.x-tagGroups[].name'  $REDOC_ROOT_API_PATH | sort)

tagsDescription=$(yq e '.tags[].name'  $REDOC_ROOT_API_PATH | sort)

#diff -B <(echo $tagsDescription) <(echo $tagsPublished)

pathsTags=$(yq e '.paths[].*'  $OPENAPI_PUBLIC_SPEC_PATH | grep "tags:" -A 1 | grep '^ *-' | tr -d '-' | sort | uniq)

diff -B <(echo $tagsPublished) <(echo $pathsTags) || exit 1

spaceSeparatedPathsTags=$(echo "$pathsTags" | tr '\n' ' ')
spaceSeparatedTagsDescription=$(echo "$tagsDescription" | tr '\n' ' ')
spaceSeparatedParentGroupTags=$(echo "$parentGroupTags" | tr '\n' ' ')

for tag in $spaceSeparatedTagsDescription; do
    # Skip parent group tags - they are category headers for x-parent hierarchy (GitBook)
    isParentGroup=false
    for parentTag in $spaceSeparatedParentGroupTags; do
        if [ "$tag" = "$parentTag" ]; then
            isParentGroup=true
            break
        fi
    done
    if [ $isParentGroup = true ]; then
        continue
    fi

    found=false
    for pathTag in $spaceSeparatedPathsTags; do
        if [ "$tag" = "$pathTag" ]; then
            found=true
            break
        fi
    done
    if [ $found = false ]; then
        echo "Tag '$tag' in tagsDescription does not exist in pathsTags"
        exit 1
    fi
done