stages="stage1 stage2 stage3 stage4 stage5"
count=5
prev=

for stage in $stages; do
    echo "$stage"
    (cd $stage && \
        go test -run nothing -bench . -benchmem -count $count > gobench$count.txt && \
        benchstat -split XXX $prev gobench$count.txt > gobenchstat.txt \
    )
    prev=../$stage/gobench$count.txt
done
