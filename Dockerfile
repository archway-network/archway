FROM scratch

COPY ./archwayd /usr/bin/archwayd

WORKDIR /root/.archway

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

ENTRYPOINT [ "/usr/bin/archwayd" ]

CMD [ "help" ]
