#!/bin/bash

# Pastikan Anda berada di direktori yang benar
cd "$(dirname "$0")"

# Nama file CSV
CSV_FILE="incoming_order_delivery.csv"

start_time=$(date +%s%N)

# Fungsi untuk memanggil createAssets
function CreateAsset() {
    while IFS=',' read -r noResi statusPengiriman lokasiBarang operator buktiStatus
    do
        noResi=$(echo "$noResi" | tr -d '\r')
        statusPengiriman=$(echo "$statusPengiriman" | tr -d '\r')
        lokasiBarang=$(echo "$lokasiBarang" | tr -d '\r')
        operator=$(echo "$operator" | tr -d '\r')
        buktiStatus=$(echo "$buktiStatus" | tr -d '\r')

        # Lewati header
        if [[ "$noResi" != "Pengiriman" ]]; then
            # Meng-escape tanda kutip ganda
            noResi=$(echo "$noResi" | sed 's/"/\\"/g')
            statusPengiriman=$(echo "$statusPengiriman" | sed 's/"/\\"/g')
            lokasiBarang=$(echo "$lokasiBarang" | sed 's/"/\\"/g')
            operator=$(echo "$operator" | sed 's/"/\\"/g')
            buktiStatus=$(echo "$buktiStatus" | sed 's/"/\\"/g')

            echo "Creating order: $noResi, Status pengiriman: $statusPengiriman, Lokasi Barang: $lokasiBarang, Operator: $operator, Bukti status: $buktiStatus"
            # Panggil chaincode untuk createAssets
            peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C jkt-jgj -n delivery_status --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c "{\"function\":\"CreateAsset\",\"Args\":[\"$noResi\",\"$statusPengiriman\",\"$lokasiBarang\",\"$operator\",\"$buktiStatus\"]}"
        fi
    done < "$CSV_FILE"
}


# Panggil fungsi
CreateAsset

end_time=$(date +%s%N)

elapsed_time=$(( (end_time - start_time) / 1000000 ))

echo "Elapsed Time: $elapsed_time ms"