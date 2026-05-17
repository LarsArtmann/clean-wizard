FROM gcr.io/distroless/static-debian13:nonroot

COPY clean-wizard /clean-wizard

USER 65532:65532

ENTRYPOINT ["/clean-wizard"]
