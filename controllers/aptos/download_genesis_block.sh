if [ -z "$(ls -A $KOTAL_DATA_PATH/kotal_genesis.blob)" ]
then
    echo "downloading genesis block"
    curl https://raw.githubusercontent.com/aptos-labs/aptos-networks/main/$KOTAL_NETWORK/genesis.blob -o $KOTAL_DATA_PATH/kotal_genesis.blob
else
    echo "genesis block has been downloaded before"
fi