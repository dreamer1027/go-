package main

import (
    "os/exec"
    "fmt"
    "os"
    "path"
    "io/ioutil"
)

const cgroupMount = "sys/fs/cgroup/resourcelimit"
const cgroupMountCpu = "/sys/fs/cgroup/cpu"
const cgroupMountCpuset = "/sys/fs/cgroup/cpuset"
const cgroupMountMemory = "/sys/fs/cgroup/memory"
const cgroupMountBlkio = "/sys/fs/cgroup/blkio"
const cgroupMountNetcls = "sys/fs/cgroup/net_cls"

type processControl interface {
    Init();
    LimitRes();
    ChangeRes();
    DeleteRes();
    AddPidsLimit();
}

type resourceLimit struct{
    PID string;//进程PID

    cfs_quota_us string;//cpu时间周期内可用cpu时间
    cfs_period_us string;//cpu时间周期

    cpus string;//使用的cpu线程
    mems string;//限制内存节点

    limit_in_bytes string;//进程内存使用限制

    read_bps_device string;//IO 读速度限制
    write_bps_device string;//IO 写速度限制

    bandwidth string; //进程使用网络带宽限制
}

type resControl struct{

}

/*对Linux下cgroup和tc进行初始化*/
func (rescontrol resControl) Init() {
    /*挂载cpu subsystem到/cgroup/cpu目录（hierarchy)*/
    cmd1 := exec.Command("mount","-t","cgroup","-o","cpu","cpu","/cgroup/cpu")
    cmd1.Stdin = os.Stdin
    cmd1.Stdout = os.Stdout
    cmd1.Stderr = os.Stderr
    if err := cmd1.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(1)
    }
    /*挂载cpuset subsystem到/cgroup/cpuset目录（hierarchy)*/
    cmd2 := exec.Command("mount","-t","cgroup","-o","cpuset","cpuset","/cgroup/cpuset")
    cmd2.Stdin = os.Stdin
    cmd2.Stdout = os.Stdout
    cmd2.Stderr = os.Stderr
    if err := cmd2.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(2)
    }
    /*挂载memory subsystem到/cgroup/memory目录（hierarchy)*/
    cmd3 := exec.Command("mount","-t","cgroup","-o","memory","memory","/cgroup/memory")
    cmd3.Stdin = os.Stdin
    cmd3.Stdout = os.Stdout
    cmd3.Stderr = os.Stderr
    if err := cmd3.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(3)
    }
    /*挂载blkio subsystem到/cgroup/blkio目录（hierarchy)*/
    cmd4 := exec.Command("mount","-t","cgroup","-o","blkio","blkio","/cgroup/blkio")//挂载subsystem
    cmd4.Stdin = os.Stdin
    cmd4.Stdout = os.Stdout
    cmd4.Stderr = os.Stderr
    if err := cmd4.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(4)
    }
    /*挂载net_cls subsystem到/cgroup/net_cls目录（hierarchy)*/
    cmd5 := exec.Command("mount","-t","cgroup","-o","net_cls","net_cls","/cgroup/net_cls")//挂载subsystem
    cmd5.Stdin = os.Stdin
    cmd5.Stdout = os.Stdout
    cmd5.Stderr = os.Stderr
    if err := cmd5.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(5)
    }
    /*对Linux下tc进行初始化*/

    /*在eth0上创建添加队列规则qdisc,此处使用htb流量控制*/
    cmd6 := exec.Command("tc","qdisc","add","dev","eth0","root","handle","1:htb")
    cmd6.Stdin = os.Stdin
    cmd6.Stdout = os.Stdout
    cmd6.Stderr = os.Stderr
    if err := cmd6.Run(); err != nil {
        fmt.Println("ERROR", err)
        os.Exit(6)
    }   
}

/*限制进程资源*/
func (rescontrol resControl) LimitRes(reslimit resourceLimit) int{

    /*控制进程cpu资源*/
    if reslimit.cfs_quota_us != " " && reslimit.cfs_period_us != " " {
        os.Mkdir(path.Join(cgroupMountCpu, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "tasks") , []byte(reslimit.PID) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "cpu.cfs_quota_us") , []byte("reslimit.cfs_quota_us"), 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "cpu.cfs_period_us") , []byte("reslimit.cfs_period_us"), 0644)
    }

    if reslimit.cpus != " " && reslimit.mems != " " {
        os.Mkdir(path.Join(cgroupMountCpuset, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "tasks") , []byte(reslimit.PID) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "cpuset.cpus") , []byte("reslimit.cpus"), 0644)
        ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "cpuset.mems") , []byte("reslimit.mems"), 0644)
    }
    
    /*控制进程内存资源*/
    if reslimit.limit_in_bytes != " " {
        os.Mkdir(path.Join(cgroupMountMemory, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "tasks") , []byte(reslimit.PID) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "memory.limit_in_bytes") , []byte("reslimit.limit_in_bytes"), 0644)
    }

    /*控制进程IO资源*/
    if reslimit.read_bps_device != " " && reslimit.write_bps_device != " " {
        os.Mkdir(path.Join(cgroupMountBlkio, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountBlkio, "resourcelimit", "tasks") , []byte(reslimit.PID) , 0644)
        ioutil.WriteFile(path.Join(cgroupMountBlkio, "resourcelimit", "memory.read_bps_device") , []byte("reslimit.read_bps_device"), 0644)
        ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "memory.write_bps_device") , []byte("reslimit.write_bps_device"), 0644)
    }
    
    /*配置cgourp模块控制进程带宽资源（同时需对tc模块进行配置）*/
    if reslimit.bandwidth != " " {
        os.Mkdir(path.Join(cgroupMountNetcls, "resourceLimit"), 0755)
        ioutil.WriteFile(path.Join(cgroupMountNetcls, "resourcelimit", "net_cls.classid") , []byte("0x100001"), 0644)

        /*在qdisc下建立分类，classid为1:1，与写入cgroup/net_cls.classid中一致*/
        cmd1 := exec.Command("tc","class","add", "dev","eth0","parent","1:classid","1:1","htb","rate",resourcelimit.bandwidth)
        cmd1.Stdin = os.Stdin
        cmd1.Stdout = os.Stdout
        cmd1.Stderr = os.Stderr
        if err := cmd1.Run(); err != nil {
            fmt.Println("ERROR", err)
            os.Exit(7)
        }

        /*添加cgroup过滤器*/
        cmd2 := exec.Command("tc","filter","add","dev","eth0","parent","1:protocol","ip","prio","10","handle","1:","cgroup")
        cmd2.Stdin = os.Stdin
        cmd2.Stdout = os.Stdout
        cmd2.Stderr = os.Stderr
        if err := cmd2.Run(); err != nil {
            fmt.Println("ERROR", err)
            os.Exit(8)
        }
    }
    return 0
}

/*对应的资源号*/
const (
    CFS_QUOTA_US = 1     
    CFS_PERIOD_US = 2    
    CPUS = 3             
    MEMS = 4             
    LIMIT_IN_BYTES = 5   
    READ_BPS_DEVICE = 6
    WRITE_BPS_DEVICE = 7 
    BANDWIDTH = 8
)

/*对进程资源限制进行修改，传入要修改资源号的切片*/
func (rescontrol resControl) ChangeRes(flags []int, reslimit resourceLimit) int{ 
    for i := 0; i < len(flags); i++ {   
        switch flags[i] {
            case CFS_QUOTA_US:
                ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "cpu.cfs_quota_us") , []byte("reslimit.cfs_quota_us"), 0644)    
            case CFS_PERIOD_US:
                ioutil.WriteFile(path.Join(cgroupMountCpu, "resourcelimit", "cpu.cfs_period_us") , []byte("reslimit.cfs_period_us"), 0644)
            case CPUS:
                ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "cpuset.cpus") , []byte("reslimit.cpus"), 0644)
            case MEMS:
                ioutil.WriteFile(path.Join(cgroupMountCpuset, "resourcelimit", "cpuset.mems") , []byte("reslimit.mems"), 0644)
            case LIMIT_IN_BYTES:
                ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "memory.limit_in_bytes") , []byte("reslimit.limit_in_bytes"), 0644)
            case READ_BPS_DEVICE:
                ioutil.WriteFile(path.Join(cgroupMountBlkio, "resourcelimit", "memory.read_bps_device") , []byte("reslimit.read_bps_device"), 0644)
            case WRITE_BPS_DEVICE:
                ioutil.WriteFile(path.Join(cgroupMountMemory, "resourcelimit", "memory.write_bps_device") , []byte("reslimit.write_bps_device"), 0644)
            case BANDWIDTH:
                cmd := exec.Command("tc","class","change","dev","br0","parent","10:classid","1:1","htb","rate",resourcelimit.bandwidth)
                cmd.Stdin = os.Stdin
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err != nil {
                    fmt.Println("ERROR", err)
                    os.Exit(9)
                }
            default:
                return -1
        }
    }
    return 0

}

/*模块对进程资源限制进行删除cgdelete*/
func (rescontrol resControl) DeleteRes(flags []int) int{
    for i := 0; i < len(flags); i++ {
        switch flags[i]{
            case CFS_QUOTA_US , CFS_PERIOD_US:
                cmd := exec.Command("cgdelete","-r",cgroupMountCpu)
                cmd.Stdin = os.Stdin
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err != nil {
                    fmt.Println("ERROR", err)
                    os.Exit(10)
                }   
            case CPUS , MEMS:
                cmd := exec.Command("cgdelete","-r",cgroupMountCpuset)
                cmd.Stdin = os.Stdin
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err != nil {
                    fmt.Println("ERROR", err)
                    os.Exit(11)
                }
            case LIMIT_IN_BYTES:
                cmd := exec.Command("cgdelete","-r",cgroupMountMemory)
                cmd.Stdin = os.Stdin
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err != nil {
                    fmt.Println("ERROR", err)
                    os.Exit(12)
                }
            case READ_BPS_DEVICE , WRITE_BPS_DEVICE:
                cmd := exec.Command("cgdelete","-r",cgroupMountBlkio)
                cmd.Stdin = os.Stdin
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err != nil {
                    fmt.Println("ERROR", err)
                    os.Exit(13)
                }
            case BANDWIDTH:
                cmd := exec.Command("tc","qdisc","del","dev","eth0","root","handle","1:htb")
                cmd.Stdin = os.Stdin
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err != nil {
                    fmt.Println("ERROR", err)
                    os.Exit(14)
                }   
            default:
                return -1
        }
    }
    return 0
}

type processResourceAlloc struct{
    PID string;//进程PID
    RES []int;//进程要限制的资源号
}

/*已配置好资源组，让服务按组资源进行分配cgexec，传入进程结构体切片*/
func (rescontrol resControl) AddPidsLimit (PRA []processResourceAlloc) int{
    for i := 0; i < len(PRA); i++ {
        for j := 0; j < len(PRA[i].RES); j++ {
            switch PRA[i].RES[j]{
                case CFS_QUOTA_US , CFS_PERIOD_US:
                    cmd := exec.Command("echo",PRA[i].PID,">",path.Join(cgroupMountCpu, "resourcelimit", "tasks"))
                    cmd.Stdin = os.Stdin
                    cmd.Stdout = os.Stdout
                    cmd.Stderr = os.Stderr
                    if err := cmd.Run(); err != nil {
                        fmt.Println("ERROR", err)
                        os.Exit(15)
                    }   
                case CPUS , MEMS:
                    cmd := exec.Command("echo",PRA[i].PID,">",path.Join(cgroupMountCpuset, "resourcelimit", "tasks"))
                    cmd.Stdin = os.Stdin
                    cmd.Stdout = os.Stdout
                    cmd.Stderr = os.Stderr
                    if err := cmd.Run(); err != nil {
                        fmt.Println("ERROR", err)
                        os.Exit(16)
                    }
                case LIMIT_IN_BYTES:
                    cmd := exec.Command("echo",PRA[i].PID,">",path.Join(cgroupMountMemory, "resourcelimit", "tasks"))
                    cmd.Stdin = os.Stdin
                    cmd.Stdout = os.Stdout
                    cmd.Stderr = os.Stderr
                    if err := cmd.Run(); err != nil {
                        fmt.Println("ERROR", err)
                        os.Exit(17)
                    }
                case READ_BPS_DEVICE , WRITE_BPS_DEVICE:
                    cmd := exec.Command("echo",PRA[i].PID,">",path.Join(cgroupMountBlkio, "resourcelimit", "tasks"))
                    cmd.Stdin = os.Stdin
                    cmd.Stdout = os.Stdout
                    cmd.Stderr = os.Stderr
                    if err := cmd.Run(); err != nil {
                        fmt.Println("ERROR", err)
                        os.Exit(18)
                    }
                case BANDWIDTH:
                    /*net_cls通过给packet添加classid进行流控，怎么给进程packet添加sock里的classid？*/
                default:
                    return -1
            }
        }    
    }
    return 0
}

