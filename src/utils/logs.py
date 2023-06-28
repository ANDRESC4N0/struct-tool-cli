import click

def error(msg):
    click.echo(
        click.style('==> ', fg='red') + msg
    )

def chapter(msg):
    click.echo(
        click.style('==> ', fg='green') + msg
    )

def echo(msg):
    # if not ctx.obj['ECHO']:
    #     return
    click.echo(msg)