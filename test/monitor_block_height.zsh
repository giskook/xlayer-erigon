#!/bin/zsh

# URLs for the sequencer and RPC endpoints
SEQUENCER_URL="http://localhost:8123"
RPC_URL="http://localhost:8124"

# JSON-RPC payload
PAYLOAD='{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Function to get block number from a given URL
get_block_number() {
    local url=$1
    local response=$(curl -s -X POST -H "Content-Type: application/json" --data "$PAYLOAD" "$url")
    local block_number=$(echo $response | jq -r '.result')

    # Check if block_number is valid
    if [[ $block_number == null || -z $block_number ]]; then
        echo "Invalid block number received from $url"
        return
    fi

    # Convert hex block number to decimal using printf
    echo $((16#${block_number#0x}))
}

# Monitor block heights every 5 seconds
while true; do
    sequencer_block_number=$(get_block_number $SEQUENCER_URL)
    rpc_block_number=$(get_block_number $RPC_URL)

    echo "Sequencer Block Number: $sequencer_block_number"
    echo "RPC Block Number: $rpc_block_number"

    sleep 5
done