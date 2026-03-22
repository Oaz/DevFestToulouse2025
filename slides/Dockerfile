FROM gouz/slidesk:2.15.2

WORKDIR /slidesk

# Install needed slidesk plugins
RUN slidesk plugin install @gouz/mermaid
RUN slidesk plugin install @gouz/qrcode
RUN slidesk plugin install @Oaz/plotly

# Copy slides
COPY ./main.sdf ./
COPY ./assets ./assets/
COPY ./slides ./slides/
COPY ./templates ./templates/

# Add loader for variables customization
COPY ./run.sh ./
ENTRYPOINT ["/slidesk/run.sh"]

# Launch slidesk with timers and notes
CMD ["slidesk", "-tn"]
