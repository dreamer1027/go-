package main

import(
    "os/exec"
    "fmt"
    "os"
)

const cgroupMount = "/sys/fs/cgroup"
const cgroupMountCpu = "/sys/fs/cgroup/cpu"
const cgroupMountCpuset = "/sys/fs/cgroup/cpuset"
const cgroupMountMemory = "/sys/fs/cgroup/memory"
const cgroupMountBlkio = "/sys/fs/cgroup/blkio"

/*对Linux下cgroup和tc进行初始化*/
func init_cgroup(){
    cmd := exec.Command("mount","-t","tmpfs","cgroup_root",cgroupMount)//挂载cgroup
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(1)
    }
}

func init_tc(){
	/*在eth0上建立队列*/
	cmd := exec.Command("tc",'qdisc add dev eth0 root handle 1: cbq bandwidth 10Mbit avpkt 1000 cell 8 mpu 64')
	if err = cmd.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(2)
    }
    
    /*建立分类*/
    cmd := exec.Command("tc",'class add dev eth0 parent 1:0 classid 1:1 cbq bandwidth 10Mbit maxburst 20 allot 1514 prio 1 avpkt 1000 cell 8 weight 1Mbit')
	if err = cmd.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(3)
    }
}

type Resource_Limit struct{
    PID int;//进程PID

	cfs_quota_us int;//cpu时间周期内可用cpu时间
    cfs_period_us int;//cpu时间周期

    cpus string;//使用的cpu线程
    mems string;//限制内存节点

    limit_in_bytes int32;//进程内存使用限制

    read_bps_device string;//IO 读速度限制
    write_bps_device string;//IO 写速度限制

    bandwidth int32;//进程使用网络带宽限制
}

func limitRes(reslimit Resource_Limit) int{

    if reslimit.pid == nil {
        return -1;
    }

    /*Linux内核tc模块控制带宽*/
    if reslimit.bandwidth != nil {
        ?如何根据进程号进行限速
    }
    
    /*Linux内核cgroup模块限制进程资源*/

    /*控制进程cpu资源*/
    if reslimit.cfs_quota_us != nil && reslimit.cfs_period_us != nil {
        os.Mkdir(path.Join(cgroupMountCpu, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "tasks") , []byte(reslimit.Pid) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "cpu.cfs_quota_us") , []byte("reslimit.cfs_quota_us"), 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "cpu.cfs_period_us") , []byte("reslimit.cfs_period_us"), 0644)
    }

    if reslimit.cpus != nil && reslimit.mems != nil {
        os.Mkdir(path.Join(cgroupMountCpuset, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "tasks") , []byte(reslimit.Pid) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "cpuset.cpus") , []byte("reslimit.cpus"), 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "cpuset.mems") , []byte("reslimit.mems"), 0644)
    }
    
    /*控制进程内存资源*/
    if reslimit.limit_in_bytes != nil {
        os.Mkdir(path.Join(cgroupMountMemory, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "tasks") , []byte(reslimit.Pid) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "memory.limit_in_bytes") , []byte("reslimit.limit_in_bytes"), 0644)
    }

    /*控制进程IO资源*/
    if reslimit.read_bps_device != nil && reslimit.write_bps_device != nil {
        os.Mkdir(path.Join(cgroupMountBlkio, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountBlkio, "resourcelimit", "tasks") , []byte(reslimit.Pid) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountBlkio, "resourcelimit", "memory.read_bps_device") , []byte("reslimit.read_bps_device"), 0644)
        ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "memory.write_bps_device") , []byte("reslimit.write_bps_device"), 0644)
    }
    return 0
}