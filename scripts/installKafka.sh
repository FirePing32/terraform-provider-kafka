if test -f "$HOME/.kafka/kafka.tgz"
then
    cd $HOME/.kafka
    tar -xzf kafka.tgz
else
    echo "kafka binary not found in $HOME/.kafka"
fi