# unitard - automatically deploy a systemd unit file from your application

## Synopsis

    import "github.com/tardisx/unitard"

    func main() {
        appName := "coolapp"
    	if deploy {
    		unit, _ := unitard.NewUnit(appName)
	    	unit.Deploy()
            os.Exit(0)
        }
        // rest of your application here
    }

## What it does

`Deploy()` automatically creates a systemd unit file, reloads the systemd daemon
so it can use it, enables the unit (so it starts on boot) and starts the service
running.

This means you can have a single binary deployment. Copy your executable to "somewhere"
on your target system, run it with `-deploy` (or however you have enabled the call to `Deploy()`)
and your application starts running in the background, and will restart on boot.

There is also an `Undeploy()` func, which you should of course
provide as an option to your users. It stops the running service, removes the unit file and restarts systemd.

## What's with the name?

It's the systemd UNIT for Automatic Restart Deployment. Or, just a stupid pun based on my username.

## Does this work for root?

It's designed to not. It leverages the systemd `--user` facility, where users can configure
their own services to run persistently, with all configuration being done out of their home
directory (in `~/.config/systemd`).

See https://wiki.archlinux.org/title/Systemd/User for more information.

## It works! Until I logout, and then my program stops!

You need to enable "lingering" - see the link above.

## I want it to do X

It's designed to be almost zero configuration - you just provide an application name
(which gets used to name the `.service` file). This is by intent. 

However it's not impossible that there are sensible user-configurable things. Raise an issue.