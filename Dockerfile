FROM scratch

COPY ./archwayd /usr/bin/archwayd

ARG USER_ID
ARG GROUP_ID

RUN addgroup --gid $GROUP_ID archway
RUN adduser -S -h /archway -D archway -u $USER_ID
USER archway

WORKDIR ~/.archway

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

ENTRYPOINT [ "/usr/bin/archwayd" ]

CMD [ "help" ]
